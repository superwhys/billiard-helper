package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/miebyte/goutils/utils/ptrx"
	"github.com/superwhys/billiard-helper/models/constant"
	"github.com/superwhys/billiard-helper/models/dbmodels"
	"github.com/superwhys/billiard-helper/models/errcode"
	"github.com/superwhys/billiard-helper/models/request"
	"github.com/superwhys/billiard-helper/models/response"
	"github.com/superwhys/billiard-helper/pkg/hash"
	"github.com/superwhys/billiard-helper/ports"
)

type roomService struct {
	srvCtx *ServiceContext
}

// NewRoomService 创建房间服务
func NewRoomService(srvCtx *ServiceContext) ports.RoomService {
	return &roomService{
		srvCtx: srvCtx,
	}
}

func (s *roomService) CreateRoom(ctx context.Context, req *request.CreateRoomRequest) error {
	if req == nil {
		return errcode.ErrCodeInvalidRequest
	}

	return s.srvCtx.RoomRepo.CreateRoom(ctx, req)
}

func (s *roomService) GetRoom(ctx context.Context, req *request.GetRoomRequest) (*response.Room, error) {
	if req == nil {
		return nil, errcode.ErrCodeInvalidRequest
	}

	room, err := s.srvCtx.RoomRepo.GetRoom(ctx, req.RoomID)
	if err != nil {
		return nil, err
	}

	return &response.Room{Room: room.ToType()}, nil
}

func (s *roomService) GetUserRooms(ctx context.Context, req *request.GetUserRoomsRequest) ([]*response.Room, error) {
	if req == nil {
		return nil, errcode.ErrCodeInvalidRequest
	}

	rooms, err := s.srvCtx.RoomRepo.GetUserRooms(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	resp := make([]*response.Room, 0, len(rooms))
	for _, room := range rooms {
		resp = append(resp, &response.Room{Room: room.ToType()})
	}
	return resp, nil
}

func (s *roomService) JoinRoom(ctx context.Context, req *request.JoinRoomRequest) (*response.Player, error) {
	if req == nil {
		return nil, errcode.ErrCodeInvalidRequest
	}

	// 检查房间是否存在
	_, err := s.srvCtx.RoomRepo.GetRoom(ctx, req.RoomID)
	if err != nil {
		return nil, err
	}

	// 生成玩家幂等 code
	code := hash.GenerateHash(
		fmt.Sprintf("%d", req.RoomID),
		fmt.Sprintf("%d", req.UserID),
		fmt.Sprintf("%d", req.PlayerType),
		fmt.Sprintf("%s", req.PlayerNickName),
	)

	// 检查玩家是否已经加入房间
	player, err := s.srvCtx.PlayerRepo.GetPlayerByCode(ctx, code)
	if err != nil && !errors.Is(err, errcode.ErrCodePlayerNotFound) {
		return nil, err
	}
	if player != nil {
		return nil, errcode.ErrCodePlayerAlreadyJoined
	}

	// 创建玩家记录
	playerModel := &dbmodels.Player{
		RoomID:   req.RoomID,
		UserID:   ptrx.Uint(req.UserID),
		NickName: req.PlayerNickName,
		Avatar:   "",
		Type:     req.PlayerType,
	}
	err = s.srvCtx.PlayerRepo.CreatePlayer(ctx, playerModel)
	if err != nil {
		return nil, err
	}

	err = s.srvCtx.Socket.Of(constant.BilliardNamespace).To(s.socketRoomID(req.RoomID)).Emit(constant.EventJoinRoom, playerModel.ToType())
	if err != nil {
		return nil, fmt.Errorf("emit join room event failed: %w", err)
	}

	return &response.Player{Player: playerModel.ToType()}, nil
}

func (s *roomService) LeaveRoom(ctx context.Context, req *request.LeaveRoomRequest) error {
	if req == nil {
		return errcode.ErrCodeInvalidRequest
	}

	err := s.srvCtx.PlayerRepo.DeletePlayer(ctx, req.PlayerCode)
	if err != nil {
		return fmt.Errorf("delete player failed: %w", err)
	}

	err = s.srvCtx.Socket.Of(constant.BilliardNamespace).To(s.socketRoomID(req.RoomID)).Emit(constant.EventLeaveRoom, req.PlayerCode)
	if err != nil {
		return fmt.Errorf("emit leave room event failed: %w", err)
	}
	return nil
}

func (s *roomService) DeleteRoom(ctx context.Context, req *request.DeleteRoomRequest) error {
	if req == nil {
		return errcode.ErrCodeInvalidRequest
	}

	if err := s.srvCtx.RoomRepo.DeleteRoom(ctx, req.RoomID); err != nil {
		return err
	}
	return nil
}

func (s *roomService) socketRoomID(roomID uint) string {
	return fmt.Sprintf("room_%d", roomID)
}
