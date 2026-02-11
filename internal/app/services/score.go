package services

import (
	"context"
	"fmt"

	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/app/factory"
	"github.com/superwhys/billiard-helper/internal/constant"
	"github.com/superwhys/billiard-helper/internal/domain/event"
	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/domain/shared"
	"github.com/superwhys/billiard-helper/internal/errcode"
	"gorm.io/datatypes"
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

	if !match.IsStart() {
		return nil, errcode.ErrCodeMatchNotInProgress
	}

	if match.OwnerID != userID {
		return nil, errcode.ErrCodeMatchNotFound
	}

	if match.MatchRound != round {
		return nil, errcode.ErrCodeMatchRoundNotMatch
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
	match, err := a.findMatch(ctx, req.UserID, req.MatchID, req.Round)
	if err != nil {
		return nil, fmt.Errorf("find match failed: %w", err)
	}

	matchGame, err := a.findMatchGame(ctx, match.ID, req.Round)
	if err != nil {
		return nil, fmt.Errorf("find match game failed: %w", err)
	}

	var newScores map[string]any
	err = a.repoFactory.WithTransaction(ctx, func(factory factory.IRepoFactory) error {
		eventRepo := factory.EventRepo()
		matchGameRepo := factory.MatchGameRepo()

		// 添加分数同步事件
		event := &event.Event{
			MatchID:    match.ID,
			OperatorID: req.UserID,
			Round:      req.Round,
			EventType:  constant.EventPlayerScoreSync,
			Data: datatypes.JSONMap{
				"score_actions": req.ScoreActions,
				"context":       req.Context,
			},
		}
		err := eventRepo.AddEvent(ctx, event)
		if err != nil {
			return fmt.Errorf("add score sync event failed: %w", err)
		}

		// TODO: 计算并修改当前比赛轮次中的分数快照
		// 计算分数快照的逻辑：
		// 1. 获取当前的分数快照
		// 2. 遍历当前事件的 ScoreActions，根据 ScoreActions 中的 PlayerIds 和 Score 计算新的分数快照
		//     2.1 这里也许要使用策略模式，不同的玩法有不同的分数计算逻辑
		//     2.2 然后这里传递当前房间的玩家，当前的分数快照，以及分数事件数据，来计算新的分数快照
		// 3. 更新分数快照

		matchGame.LastEventID = event.ID
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
	matchGameRepo := a.repoFactory.MatchGameRepo()
	eventRepo := a.repoFactory.EventRepo()

	matchGame, err := a.findMatchGame(ctx, req.MatchID, req.Round)
	if err != nil {
		return nil, fmt.Errorf("find match game failed: %w", err)
	}

	lastEventID := matchGame.LastEventID
	deletedEvent, err := eventRepo.DeleteEvent(ctx, lastEventID)
	if err != nil {
		return nil, fmt.Errorf("delete event failed: %w", err)
	}

	// 撤回该时间的操作
	_ = deletedEvent

	lastEvent, err := eventRepo.GetLastEvent(ctx, req.MatchID, req.Round)
	if err != nil {
		return nil, fmt.Errorf("get last event failed: %w", err)
	}

	// TODO: 修改当前比赛轮次中的分数快照
	matchGame.LastEventID = lastEvent.ID
	err = matchGameRepo.Update(ctx, matchGame)
	if err != nil {
		return nil, fmt.Errorf("update match game failed: %w", err)
	}

	return newScores, nil
}

func (a *ScoreApp) ListScores(ctx context.Context, req *dto.MatchScoreListReq) ([]*dto.MatchScoreSyncEvent, error) {
	return nil, nil
}
