package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/superwhys/billiard-helper/api/middlewares"
	"github.com/superwhys/billiard-helper/internal/comet/queue"
	"github.com/superwhys/billiard-helper/internal/dal/cache"
	"github.com/superwhys/billiard-helper/internal/models/constant"
	"github.com/superwhys/billiard-helper/internal/models/dbmodels"
	"github.com/superwhys/billiard-helper/internal/models/errcode"
	"github.com/superwhys/billiard-helper/internal/models/request"
	"github.com/superwhys/billiard-helper/internal/models/response"
	"github.com/superwhys/billiard-helper/internal/models/types"
	"github.com/superwhys/billiard-helper/internal/ports"
	"gorm.io/gorm"
)

type scoresService struct {
	srvCtx *ServiceContext
}

func NewScoresService(srvCtx *ServiceContext) ports.ScoresService {
	return &scoresService{
		srvCtx: srvCtx,
	}
}

func (s *scoresService) AddScore(ctx context.Context, req *request.AddScoreRequest) error {
	if req == nil {
		return errcode.ErrCodeInvalidRequest
	}

	return s.changeScore(ctx, req.RoomID, req.PlayerID, req.OperatorID, req.Change, constant.EventPlayerScoreAdd)
}

func (s *scoresService) MinusScore(ctx context.Context, req *request.MinusScoreRequest) error {
	if req == nil {
		return errcode.ErrCodeInvalidRequest
	}

	return s.changeScore(ctx, req.RoomID, req.PlayerID, req.OperatorID, -req.Change, constant.EventPlayerScoreMinus)
}

func (s *scoresService) ResetScore(ctx context.Context, req *request.ResetScoreRequest) error {
	if req == nil {
		return errcode.ErrCodeInvalidRequest
	}

	// Reset 特殊处理：Change 值取决于当前总分
	return s.handleResetScore(ctx, req)
}

func (s *scoresService) UndoScore(ctx context.Context, req *request.UndoScoreRequest) error {
	if req == nil {
		return errcode.ErrCodeInvalidRequest
	}

	userClaims, err := middlewares.TokenClaimsFromContext(ctx)
	if err != nil {
		return errcode.ErrCodeNoToken
	}

	// 分布式锁
	rdb := s.srvCtx.RedisClient
	roomLock := cache.RoomLockCache(fmt.Sprintf("room:%d", req.RoomID))
	if err := roomLock.Lock(ctx, rdb); err != nil {
		return err
	}
	defer roomLock.Unlock(ctx, rdb)

	// 校验房间是否存在
	exist, err := s.srvCtx.RoomRepo.IsRoomExist(ctx, req.RoomID)
	if err != nil {
		return err
	}
	if !exist {
		return errcode.ErrCodeRoomNotFound
	}

	// 获取最近一条记录
	latestScore, err := s.srvCtx.ScoreRepo.GetLatestScore(ctx, req.RoomID, req.PlayerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrCodeUndoScoreFailed
		}
		return err
	}

	if latestScore == nil {
		return errcode.ErrCodeUndoScoreFailed
	}

	// 计算回滚后的 Change 和 Total
	// 假设上一条是 Add(+5), Total=10 (Previous=5)
	// 撤回就是: Change = -5, Total = 10 - 5 = 5
	undoChange := -latestScore.Change
	newTotal := latestScore.Total + undoChange // 也就是 latestScore.Total - latestScore.Change

	// 创建撤回记录
	score := &dbmodels.Scores{
		RoomID:     req.RoomID,
		PlayerID:   req.PlayerID,
		OperatorID: req.OperatorID,
		Change:     undoChange,
		Total:      newTotal,
	}

	if err := s.srvCtx.ScoreRepo.CreateScore(ctx, score); err != nil {
		return err
	}

	// 补充 Operator 信息
	operator, err := s.srvCtx.PlayerRepo.GetPlayerByID(ctx, req.OperatorID)
	if err == nil && operator != nil {
		score.Operator = operator
	}

	// 推送消息
	return s.publishScoreEvent(ctx, req.RoomID, userClaims.User.ID, userClaims.SessionID, constant.EventPlayerScoreUndo, score.ToType())
}

func (s *scoresService) GetRoomScores(ctx context.Context, req *request.GetRoomScoresRequest) ([]*response.Scores, error) {
	if req == nil {
		return nil, errcode.ErrCodeInvalidRequest
	}

	scores, err := s.srvCtx.ScoreRepo.GetRoomScores(ctx, req.RoomID)
	if err != nil {
		return nil, err
	}

	resp := make([]*response.Scores, 0, len(scores))
	for _, score := range scores {
		resp = append(resp, &response.Scores{Scores: score.ToType()})
	}
	return resp, nil
}

// changeScore 处理通用的分数变更
func (s *scoresService) changeScore(ctx context.Context, roomID, playerID, operatorID uint, change int, event string) error {
	userClaims, err := middlewares.TokenClaimsFromContext(ctx)
	if err != nil {
		return errcode.ErrCodeNoToken
	}

	// 分布式锁
	rdb := s.srvCtx.RedisClient
	roomLock := cache.RoomLockCache(fmt.Sprintf("room:%d", roomID))
	if err := roomLock.Lock(ctx, rdb); err != nil {
		return err
	}
	defer roomLock.Unlock(ctx, rdb)

	// 校验房间是否存在
	exist, err := s.srvCtx.RoomRepo.IsRoomExist(ctx, roomID)
	if err != nil {
		return err
	}
	if !exist {
		return errcode.ErrCodeRoomNotFound
	}

	// 获取当前总分
	currentTotal := 0
	latestScore, err := s.srvCtx.ScoreRepo.GetLatestScore(ctx, roomID, playerID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if latestScore != nil {
		currentTotal = latestScore.Total
	}

	newTotal := currentTotal + change

	// 创建记录
	score := &dbmodels.Scores{
		RoomID:     roomID,
		PlayerID:   playerID,
		OperatorID: operatorID,
		Change:     change,
		Total:      newTotal,
	}

	if err := s.srvCtx.ScoreRepo.CreateScore(ctx, score); err != nil {
		return err
	}

	// 补充 Operator 信息用于推送
	// 为了前端展示方便，通常最好带上 Operator 的信息
	operator, err := s.srvCtx.PlayerRepo.GetPlayerByID(ctx, operatorID)
	if err == nil && operator != nil {
		score.Operator = operator
	}

	// 推送消息
	return s.publishScoreEvent(ctx, roomID, userClaims.User.ID, userClaims.SessionID, event, score.ToType())
}

// handleResetScore 处理重置逻辑
func (s *scoresService) handleResetScore(ctx context.Context, req *request.ResetScoreRequest) error {
	userClaims, err := middlewares.TokenClaimsFromContext(ctx)
	if err != nil {
		return errcode.ErrCodeNoToken
	}

	rdb := s.srvCtx.RedisClient
	roomLock := cache.RoomLockCache(fmt.Sprintf("room:%d", req.RoomID))
	if err := roomLock.Lock(ctx, rdb); err != nil {
		return err
	}
	defer roomLock.Unlock(ctx, rdb)

	exist, err := s.srvCtx.RoomRepo.IsRoomExist(ctx, req.RoomID)
	if err != nil {
		return err
	}
	if !exist {
		return errcode.ErrCodeRoomNotFound
	}

	latestScore, err := s.srvCtx.ScoreRepo.GetLatestScore(ctx, req.RoomID, req.PlayerID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	currentTotal := 0
	if latestScore != nil {
		currentTotal = latestScore.Total
	}

	// 如果已经是0，是否还需要记录？通常为了审计，还是记录一下 Reset 操作，Change = 0
	change := -currentTotal

	score := &dbmodels.Scores{
		RoomID:     req.RoomID,
		PlayerID:   req.PlayerID,
		OperatorID: req.OperatorID,
		Change:     change,
		Total:      0,
	}

	if err := s.srvCtx.ScoreRepo.CreateScore(ctx, score); err != nil {
		return err
	}

	// 补充 Operator
	op, err := s.srvCtx.PlayerRepo.GetPlayerByID(ctx, req.OperatorID) // 需要确认 Repo 有此方法
	if err == nil {
		score.Operator = op
	}

	return s.publishScoreEvent(ctx, req.RoomID, userClaims.User.ID, userClaims.SessionID, constant.EventPlayerScoreReset, score.ToType())
}

func (s *scoresService) publishScoreEvent(ctx context.Context, roomID uint, userID uint, sessionID string, event string, score *types.Scores) error {
	msg := &constant.ScoreUpdateMessage{
		EventMsgBase: constant.EventMsgBase{
			UserID:    userID,
			RoomID:    types.SocketRoomID(roomID),
			SessionID: sessionID,
		},
		Score: score,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal score msg failed: %w", err)
	}

	queueMsg := &queue.QueueMessage{
		Event: event,
		Data:  bytes,
	}

	return s.srvCtx.EventQueue.Publish(ctx, constant.BilliardMessageChannel, queueMsg)
}
