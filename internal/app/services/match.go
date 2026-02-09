package services

import (
	"context"
	"encoding/json"

	"github.com/miebyte/goutils/utils/ptrx"
	"github.com/superwhys/billiard-helper/internal/app/assembler"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/app/factory"
	"github.com/superwhys/billiard-helper/internal/constant"
	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/domain/shared"
	"github.com/superwhys/billiard-helper/internal/errcode"
	"github.com/superwhys/billiard-helper/internal/infra/cache"
)

type MatchApp struct {
	repoFactory    factory.IRepoFactory
	serviceFactory *factory.DomainServiceFactory
	matchAssembler *assembler.MatchAssembler
	eventBus       shared.EventBus
	lockManager    *cache.LockManager
}

func NewMatchApp(
	serviceFactory *factory.DomainServiceFactory,
	repoFactory factory.IRepoFactory,
	eventBus shared.EventBus,
	lockManager *cache.LockManager,
) *MatchApp {
	return &MatchApp{
		repoFactory:    repoFactory,
		serviceFactory: serviceFactory,
		eventBus:       eventBus,
		lockManager:    lockManager,
		matchAssembler: assembler.NewMatchAssembler(),
	}
}

// CreateMatch 创建比赛
func (a *MatchApp) CreateMatch(ctx context.Context, req *dto.CreateMatchRequest) (*dto.Match, error) {
	matchEntity := a.matchAssembler.CreateMatchReqToMatch(req)

	err := a.repoFactory.WithTransaction(ctx, func(repoFactory factory.IRepoFactory) error {
		matchService := a.serviceFactory.MatchService(repoFactory)
		err := matchService.CreateMatch(ctx, req.UserID, matchEntity)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return a.matchAssembler.ToMatchDTO(matchEntity), nil
}

func (a *MatchApp) UpdateMatch(ctx context.Context, req *dto.UpdateMatchRequest) (*dto.Match, error) {
	matchRepo := a.repoFactory.MatchRepo()

	m, err := matchRepo.FindByID(ctx, req.MatchID, false)
	if err != nil {
		return nil, err
	}
	if m.OwnerID != req.UserID {
		return nil, errcode.ErrCodeMatchNotFound
	}

	if m.Status != match.MatchStatusPending {
		return nil, errcode.ErrCodeMatchNotPending
	}

	m.Name = req.Name
	m.Config.TargetScore = req.TargetScore

	err = matchRepo.Update(ctx, m)
	if err != nil {
		return nil, err
	}

	return a.matchAssembler.ToMatchDTO(m), nil
}

// JoinRoom 加入房间
func (a *MatchApp) JoinMatch(ctx context.Context, req *dto.JoinMatchRequest) (*dto.Match, error) {
	matchRepo := a.repoFactory.MatchRepo()
	matchService := a.serviceFactory.MatchService(a.repoFactory)

	// 1. 获取比赛锁
	lock := a.lockManager.MatchLock(req.MatchID)
	if err := lock.Lock(ctx); err != nil {
		return nil, err
	}
	defer lock.Unlock(ctx)

	// 2. 获取比赛房间
	matchRoom, err := matchRepo.FindByID(ctx, req.MatchID, true)
	if err != nil {
		return nil, err
	}

	// 3. 创建玩家
	player := match.NewPlayer(req.MatchID, ptrx.Uint(req.UserID), req.NickName, req.PlayerType)

	// 4. 加入比赛
	joinedPlayer, err := matchService.JoinMatch(ctx, matchRoom, player)
	if err != nil {
		return nil, err
	}

	// 5. 发送事件
	playerDTO := a.matchAssembler.ToPlayerDTO(joinedPlayer)
	matchDTO := a.matchAssembler.ToMatchDTO(matchRoom)
	msg := dto.JoinMatchEventMessage{
		EventMsgBase: dto.EventMsgBase{
			UserID:  req.UserID,
			MatchID: matchDTO.ID,
		},
		Player: &playerDTO,
	}
	_ = a.publishEvent(ctx, constant.EventPlayerJoinRoom, msg)

	return matchDTO, nil
}

// StartRoom 开始比赛
func (a *MatchApp) StartMatch(ctx context.Context, req *dto.MatchActionRequest) error {
	matchRepo := a.repoFactory.MatchRepo()
	matchService := a.serviceFactory.MatchService(a.repoFactory)

	// 1. 获取比赛锁
	lock := a.lockManager.MatchLock(req.MatchID)
	if err := lock.Lock(ctx); err != nil {
		return err
	}
	defer lock.Unlock(ctx)

	// 2. 检查比赛是否存在
	matchRoom, err := matchRepo.FindByID(ctx, req.MatchID, false)
	if err != nil {
		return err
	}

	// 3. 开始比赛
	err = matchService.StartMatch(ctx, matchRoom)
	if err != nil {
		return err
	}

	// 4. 发布开始事件
	msg := dto.MatchStartedEventMessage{
		EventMsgBase: dto.EventMsgBase{
			UserID:  req.UserID,
			MatchID: req.MatchID,
		},
	}
	_ = a.publishEvent(ctx, constant.EventMatchStarted, msg)
	return nil
}

// EndRoom 结束比赛
func (a *MatchApp) EndMatch(ctx context.Context, req *dto.MatchActionRequest) error {
	matchRepo := a.repoFactory.MatchRepo()
	matchService := a.serviceFactory.MatchService(a.repoFactory)

	// 1. 获取比赛锁
	lock := a.lockManager.MatchLock(req.MatchID)
	if err := lock.Lock(ctx); err != nil {
		return err
	}
	defer lock.Unlock(ctx)

	// 2. 检查比赛是否存在
	matchRoom, err := matchRepo.FindByID(ctx, req.MatchID, false)
	if err != nil {
		return err
	}

	// 3. 开始比赛
	err = matchService.EndMatch(ctx, matchRoom)
	if err != nil {
		return err
	}

	// 4. 发布开始事件
	msg := dto.MatchEndedEventMessage{
		EventMsgBase: dto.EventMsgBase{
			UserID:  req.UserID,
			MatchID: req.MatchID,
		},
	}
	_ = a.publishEvent(ctx, constant.EventMatchEnded, msg)
	return nil
}

// LeaveMatch 离开比赛房间
func (a *MatchApp) LeaveMatch(ctx context.Context, req *dto.MatchActionRequest) error {
	matchRepo := a.repoFactory.MatchRepo()
	playerRepo := a.repoFactory.PlayerRepo()
	matchService := a.serviceFactory.MatchService(a.repoFactory)

	// 1. 获取比赛锁
	lock := a.lockManager.MatchLock(req.MatchID)
	if err := lock.Lock(ctx); err != nil {
		return err
	}
	defer lock.Unlock(ctx)

	// 2. 获取玩家信息
	player, err := playerRepo.FindByCode(ctx, req.PlayerCode)
	if err != nil {
		return err
	}

	// 3. 获取房间信息
	matchRoom, err := matchRepo.FindByID(ctx, player.MatchID, false)
	if err != nil {
		return err
	}

	// 4. 退出比赛
	err = matchService.LeaveMatch(ctx, matchRoom, player)
	if err != nil {
		return err
	}

	// 5. 发布离开事件
	_ = a.publishEvent(ctx, constant.EventPlayerLeaveRoom, &dto.LeaveMatchEventMessage{
		EventMsgBase: dto.EventMsgBase{
			UserID:  req.UserID,
			MatchID: matchRoom.ID,
		},
		PlayerCode: req.PlayerCode,
	})

	return nil
}

// KickPlayer 踢人
func (a *MatchApp) KickMatchPlayer(ctx context.Context, req *dto.KickPlayerRequest) error {
	matchService := a.serviceFactory.MatchService(a.repoFactory)
	playerRepo := a.repoFactory.PlayerRepo()
	matchRepo := a.repoFactory.MatchRepo()

	// 1. 获取比赛锁
	lock := a.lockManager.MatchLock(req.MatchID)
	if err := lock.Lock(ctx); err != nil {
		return err
	}
	defer lock.Unlock(ctx)

	// 2. 获取玩家信息
	player, err := playerRepo.FindByCode(ctx, req.PlayerCode)
	if err != nil {
		return err
	}

	// 3. 获取房间信息
	matchRoom, err := matchRepo.FindByID(ctx, player.MatchID, false)
	if err != nil {
		return err
	}

	// 4. 踢出玩家
	err = matchService.KickMatchPlayer(ctx, matchRoom, player)
	if err != nil {
		return err
	}

	// 5. 发布踢人事件 (通常复用离开事件)
	_ = a.publishEvent(ctx, constant.EventPlayerLeaveRoom, &dto.LeaveMatchEventMessage{
		EventMsgBase: dto.EventMsgBase{
			UserID:  req.UserID,
			MatchID: matchRoom.ID,
		},
		PlayerCode: req.PlayerCode,
	})
	return nil
}

// publishEvent 辅助方法：发布消息到 EventBus
func (a *MatchApp) publishEvent(ctx context.Context, eventType string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := &shared.QueueMessage{
		Event: eventType,
		Data:  data,
	}

	return a.eventBus.Publish(ctx, constant.BilliardEventChannel, msg)
}

func (a *MatchApp) ListMatches(ctx context.Context, userId uint, req *dto.MatchListRequest) ([]*dto.Match, error) {
	matchRepo := a.repoFactory.MatchRepo()
	matches, err := matchRepo.ListMatches(ctx, string(req.MatchType), req.Limit, req.Cursor)
	if err != nil {
		return nil, err
	}

	matchDTOs := make([]*dto.Match, 0, len(matches))
	for _, match := range matches {
		matchDTOs = append(matchDTOs, a.matchAssembler.ToMatchDTO(match))
	}
	return matchDTOs, nil
}

func (a *MatchApp) GetMatchDetail(ctx context.Context, userId uint, matchID uint) (*dto.Match, error) {
	matchRepo := a.repoFactory.MatchRepo()
	match, err := matchRepo.FindByID(ctx, matchID, true)
	if err != nil {
		return nil, err
	}

	if match.OwnerID != userId {
		return nil, errcode.ErrCodeMatchNotFound
	}

	return a.matchAssembler.ToMatchDTO(match), nil
}

func (a *MatchApp) DeleteMatch(ctx context.Context, req *dto.DeleteMatchRequest) error {
	m, err := a.repoFactory.MatchRepo().FindByID(ctx, req.MatchID, true)
	if err != nil {
		return err
	}

	if m.OwnerID != req.UserID {
		return errcode.ErrCodeMatchNotFound
	}

	if m.Status != match.MatchStatusPending {
		return errcode.ErrCodeMatchNotPending
	}

	return a.repoFactory.WithTransaction(ctx, func(factory factory.IRepoFactory) error {
		matchRepo := factory.MatchRepo()
		playerRepo := factory.PlayerRepo()

		err = matchRepo.Delete(ctx, req.MatchID)
		if err != nil {
			return err
		}

		for _, player := range m.Players {
			err = playerRepo.Delete(ctx, player.ID)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
