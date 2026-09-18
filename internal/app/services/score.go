package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/miebyte/goutils/utils/ptrx"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/app/factory"
	"github.com/superwhys/billiard-helper/internal/constant"
	"github.com/superwhys/billiard-helper/internal/domain/event"
	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/domain/shared"
	"github.com/superwhys/billiard-helper/internal/errcode"
	"gorm.io/gorm"
)

type ScoreApp struct {
	serviceFactory *factory.DomainServiceFactory
	repoFactory    factory.IRepoFactory
	eventBus       shared.EventBus
}

func NewScoreApp(
	serviceFactory *factory.DomainServiceFactory,
	repoFactory factory.IRepoFactory,
	eventBus shared.EventBus,
) *ScoreApp {
	return &ScoreApp{
		serviceFactory: serviceFactory,
		repoFactory:    repoFactory,
		eventBus:       eventBus,
	}
}

func (a *ScoreApp) findMatch(ctx context.Context, userID uint, matchID uint, round uint) (*match.Match, error) {
	matchRepo := a.repoFactory.MatchRepo()
	match, err := matchRepo.FindByID(ctx, matchID, true)
	if err != nil {
		return nil, err
	}
	if err := match.AssertScoreRequest(userID, round); err != nil {
		return nil, err
	}

	return match, nil
}

func (a *ScoreApp) findMatchGame(ctx context.Context, matchID uint, round uint) (*match.MatchGame, error) {
	matchGameRepo := a.repoFactory.MatchGameRepo()
	matchGame, err := matchGameRepo.FindByMatchID(ctx, matchID, round)
	if err != nil {
		return nil, err
	}

	return matchGame, nil
}

// SyncScore 同步每次操作的分数变化
// 每次操作的分数变化都会记录一条分数事件(有可能一次操作会有多次操作变化)
// 每条操作事件都会记录到数据库中，方便后续查询和统计
// 同时，还会修改当前比赛轮次中的分数快照并且发布一个事件
func (a *ScoreApp) SyncScore(ctx context.Context, req *dto.MatchScoreSyncEvent) (map[string]any, error) {
	m, err := a.findMatch(ctx, req.UserID, req.MatchID, req.Round)
	if err != nil {
		return nil, fmt.Errorf("find match failed: %w", err)
	}

	matchGame, err := a.findMatchGame(ctx, m.ID, req.Round)
	if err != nil {
		return nil, fmt.Errorf("find match game failed: %w", err)
	}

	// 序列化分数事件数据
	scoreEventData := &event.EventData[any]{
		ScoreActions: req.ScoreActions,
		Context:      req.Context,
	}
	scoreEventDataJSON, err := json.Marshal(scoreEventData)
	if err != nil {
		return nil, fmt.Errorf("marshal score event data failed: %w", err)
	}

	event := &event.Event{
		MatchID:    m.ID,
		OperatorID: req.UserID,
		Round:      req.Round,
		EventType:  constant.EventPlayerScoreSync,
		Data:       scoreEventDataJSON,
	}

	var newScores map[string]any
	err = a.repoFactory.WithTransaction(ctx, func(factory factory.IRepoFactory) error {
		if m.MatchType == match.MatchTypeSnooker {
			m, err = factory.MatchRepo().FindByIDForUpdate(ctx, req.MatchID)
			if err != nil {
				return err
			}
			if err := m.AssertScoreRequest(req.UserID, req.Round); err != nil {
				return err
			}
			matchGame, err = factory.MatchGameRepo().FindByMatchID(ctx, req.MatchID, req.Round)
			if err != nil {
				return err
			}
			if matchGame.EndAt != 0 {
				return errcode.ErrBadRequest.WithMessage("本局已经结算")
			}
		}

		eventRepo := factory.EventRepo()
		matchGameRepo := factory.MatchGameRepo()

		// 添加分数同步事件
		err = eventRepo.AddEvent(ctx, event)
		if err != nil {
			return fmt.Errorf("add score sync event failed: %w", err)
		}

		// 计算并修改当前比赛轮次中的分数快照
		matchService := a.serviceFactory.MatchService(factory)
		newScoresJSON, err := matchService.CalculateMatchGameScore(ctx, m, matchGame, event)
		if err != nil {
			return fmt.Errorf("calculate match game score failed: %w", err)
		}

		err = json.Unmarshal(newScoresJSON, &newScores)
		if err != nil {
			return fmt.Errorf("unmarshal new scores failed: %w", err)
		}

		matchGame.Scores = newScoresJSON
		matchGame.LastEventID = new(event.ID)
		err = matchGameRepo.Update(ctx, matchGame)
		if err != nil {
			return fmt.Errorf("update match game failed: %w", err)
		}

		return nil
	})

	return newScores, err
}

func (a *ScoreApp) UndoScore(ctx context.Context, req *dto.MatchScoreUndoReq) (map[string]any, error) {
	var newScores map[string]any
	m, err := a.findMatch(ctx, req.UserID, req.MatchID, req.Round)
	if err != nil {
		return nil, fmt.Errorf("find match failed: %w", err)
	}

	matchGame, err := a.findMatchGame(ctx, req.MatchID, req.Round)
	if err != nil {
		return nil, fmt.Errorf("find match game failed: %w", err)
	}

	if matchGame.LastEventID == nil {
		return nil, errcode.ErrCodeUndoScoreFailed.WithMessage("没有可撤回的事件")
	}

	err = a.repoFactory.WithTransaction(ctx, func(factory factory.IRepoFactory) error {
		if m.MatchType == match.MatchTypeSnooker {
			m, err = factory.MatchRepo().FindByIDForUpdate(ctx, req.MatchID)
			if err != nil {
				return err
			}
			if err := m.AssertScoreRequest(req.UserID, req.Round); err != nil {
				return err
			}
			matchGame, err = factory.MatchGameRepo().FindByMatchID(ctx, req.MatchID, req.Round)
			if err != nil {
				return err
			}
			if matchGame.EndAt != 0 {
				return errcode.ErrBadRequest.WithMessage("本局已经结算")
			}
		}

		matchGameRepo := factory.MatchGameRepo()
		eventRepo := factory.EventRepo()

		if matchGame.LastEventID == nil {
			return errcode.ErrCodeUndoScoreFailed.WithMessage("没有可撤回的事件")
		}
		lastEventID := matchGame.LastEventID
		deletedEvent, err := eventRepo.DeleteEvent(ctx, ptrx.UintValue(lastEventID))
		if err != nil {
			return fmt.Errorf("delete event failed: %w", err)
		}

		matchService := a.serviceFactory.MatchService(factory)
		newScoresJSON, err := matchService.UndoMatchGameScore(ctx, m, matchGame, deletedEvent)
		if err != nil {
			return fmt.Errorf("undo match game score failed: %w", err)
		}

		err = json.Unmarshal(newScoresJSON, &newScores)
		if err != nil {
			return fmt.Errorf("unmarshal new scores failed: %w", err)
		}

		matchGame.Scores = newScoresJSON

		lastEvent, err := eventRepo.GetLastEvent(ctx, req.MatchID, req.Round)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				matchGame.LastEventID = nil
				return matchGameRepo.Update(ctx, matchGame)
			}
			return fmt.Errorf("get last event failed: %w", err)
		}

		matchGame.LastEventID = new(lastEvent.ID)
		err = matchGameRepo.Update(ctx, matchGame)
		if err != nil {
			return fmt.Errorf("update match game failed: %w", err)
		}

		return nil
	})

	return newScores, err
}

func (a *ScoreApp) ListScores(ctx context.Context, req *dto.MatchScoreListReq) ([]*dto.MatchScoreSyncEvent, error) {
	return nil, nil
}
