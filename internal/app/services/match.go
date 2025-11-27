package services

import (
	"context"
	"encoding/json"

	"github.com/superwhys/billiard-helper/internal/app/assembler"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/constant"
	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/domain/shared"
)

type MatchApp struct {
	matchService   match.IMatchService
	matchAssembler *assembler.MatchAssembler
	eventBus       shared.EventBus
}

func NewMatchApp(
	matchService match.IMatchService,
	matchAssembler *assembler.MatchAssembler,
	eventBus shared.EventBus,
) *MatchApp {
	return &MatchApp{
		matchService:   matchService,
		matchAssembler: matchAssembler,
		eventBus:       eventBus,
	}
}

// CreateRoom 创建房间
func (a *MatchApp) CreateRoom(ctx context.Context, req *dto.CreateRoomRequest) (*dto.Room, error) {
	config := a.matchAssembler.ToGameConfig(req)
	room, err := a.matchService.CreateRoom(ctx, req.UserID, config)
	if err != nil {
		return nil, err
	}
	return a.matchAssembler.ToRoomDTO(room), nil
}

// JoinRoom 加入房间
func (a *MatchApp) JoinRoom(ctx context.Context, req *dto.JoinRoomRequest) (*dto.Room, error) {
	room, err := a.matchService.JoinRoom(ctx, req.RoomID, req.UserID, req.NickName)
	if err != nil {
		return nil, err
	}

	roomDTO := a.matchAssembler.ToRoomDTO(room)
	// 发布事件
	// 注意：这里通常只需要推送给房间内的其他人，或者推送整个房间的最新状态
	// 为了简化，这里推送整个 RoomDTO，前端自己判断
	_ = a.publishEvent(ctx, constant.EventPlayerJoinRoom, roomDTO)

	return roomDTO, nil
}

// StartRoom 开始比赛
func (a *MatchApp) StartRoom(ctx context.Context, req *dto.RoomActionRequest) error {
	err := a.matchService.StartRoom(ctx, req.RoomID)
	if err != nil {
		return err
	}
	// 可以在这里发布 RoomStarted 事件
	return nil
}

// EndRoom 结束比赛
func (a *MatchApp) EndRoom(ctx context.Context, req *dto.RoomActionRequest) error {
	err := a.matchService.EndRoom(ctx, req.RoomID)
	if err != nil {
		return err
	}
	// 可以在这里发布 RoomEnded 事件
	return nil
}

// LeaveRoom 离开房间
func (a *MatchApp) LeaveRoom(ctx context.Context, req *dto.RoomActionRequest) error {
	err := a.matchService.LeaveRoom(ctx, req.RoomID, req.UserID)
	if err != nil {
		return err
	}

	// 发布离开事件
	payload := map[string]any{
		"room_id": req.RoomID,
		"user_id": req.UserID,
	}
	_ = a.publishEvent(ctx, constant.EventPlayerLeaveRoom, payload)

	return nil
}

// KickPlayer 踢人
func (a *MatchApp) KickPlayer(ctx context.Context, req *dto.KickPlayerRequest) error {
	err := a.matchService.KickPlayer(ctx, req.RoomID, req.TargetUserID)
	if err != nil {
		return err
	}

	// 发布踢人事件 (通常复用离开事件，或者有单独的 Kick 事件)
	payload := map[string]any{
		"room_id": req.RoomID,
		"user_id": req.UserID,
		"kicked":  true,
	}
	_ = a.publishEvent(ctx, constant.EventPlayerLeaveRoom, payload)
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
	// 使用默认的业务频道
	return a.eventBus.Publish(ctx, constant.BilliardMessageChannel, msg)
}
