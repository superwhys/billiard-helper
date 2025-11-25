package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/miebyte/goutils/utils/ptrx"
	"github.com/superwhys/billiard-helper/api/middlewares"
	"github.com/superwhys/billiard-helper/internal/models/constant"
	"github.com/superwhys/billiard-helper/internal/models/dbmodels"
	"github.com/superwhys/billiard-helper/internal/models/errcode"
	"github.com/superwhys/billiard-helper/internal/models/request"
	"github.com/superwhys/billiard-helper/internal/models/response"
	"github.com/superwhys/billiard-helper/internal/models/types"
	"github.com/superwhys/billiard-helper/internal/pkg/hash"
	"github.com/superwhys/billiard-helper/internal/pkg/longnet"
	"github.com/superwhys/billiard-helper/internal/ports"
	"gorm.io/gorm"
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

func (s *roomService) genRoomCode(roomID uint, userID uint, playerType types.PlayerType, playerNickName string) string {
	// 生成玩家幂等 code
	code := hash.GenerateHash(
		fmt.Sprintf("%d", roomID),
		fmt.Sprintf("%d", userID),
		fmt.Sprintf("%d", playerType),
		fmt.Sprintf("%s", playerNickName),
	)
	return code
}

func (s *roomService) CreateRoom(ctx context.Context, req *request.CreateRoomRequest) (*response.Room, error) {
	if req == nil {
		return nil, errcode.ErrCodeInvalidRequest
	}

	userClaims, err := middlewares.TokenClaimsFromContext(ctx)
	if err != nil {
		return nil, errcode.ErrCodeNoToken
	}
	userID := userClaims.User.ID

	roomModel := &dbmodels.Room{
		RoomCode: req.RoomCode,
		UserID:   userID,
		Status:   types.RoomStatusPending,
	}
	err = s.srvCtx.RoomRepo.CreateRoom(ctx, roomModel)
	if err != nil {
		return nil, err
	}
	return &response.Room{Room: roomModel.ToType()}, nil
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

	userClaims, err := middlewares.TokenClaimsFromContext(ctx)
	if err != nil {
		return nil, errcode.ErrCodeNoToken
	}
	userID := userClaims.User.ID

	rooms, err := s.srvCtx.RoomRepo.GetUserRooms(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := make([]*response.Room, 0, len(rooms))
	for _, room := range rooms {
		resp = append(resp, &response.Room{Room: room.ToType()})
	}
	return resp, nil
}

func (s *roomService) JoinRoom(ctx context.Context, req *request.JoinRoomRequest) (*response.Room, error) {
	if req == nil {
		return nil, errcode.ErrCodeInvalidRequest
	}

	userClaims, err := middlewares.TokenClaimsFromContext(ctx)
	if err != nil {
		return nil, errcode.ErrCodeNoToken
	}
	userID := userClaims.User.ID

	// 检查房间是否存在
	room, err := s.srvCtx.RoomRepo.GetRoom(ctx, req.RoomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrCodeRoomNotFound
		}
		return nil, err
	}

	roomType := room.ToType()

	// 生成玩家幂等 code
	code := s.genRoomCode(req.RoomID, userID, req.PlayerType, req.PlayerNickName)
	s.markSelfPlayer(roomType, code)

	// 确保玩家加入房间并返回玩家实例
	playerObj, err := s.ensureJoinPlayer(ctx, req, userID, code)
	if err != nil {
		return nil, err
	}

	// 获取用户 session 并加入房间
	roomID := types.SocketRoomID(req.RoomID)
	if err := s.srvCtx.SessionManager.JoinRoom(ctx, userClaims.UUID, roomID); err != nil {
		return nil, err
	}

	// 发布玩家加入房间事件
	joinMsg := &constant.JoinRoomMessage{
		UserID: userID,
		RoomID: req.RoomID,
		Player: playerObj.ToType(),
	}

	if err := s.publishRoomEvent(ctx, constant.EventPlayerJoinRoom, joinMsg); err != nil {
		return nil, fmt.Errorf("publish room event failed: %w", err)
	}

	return &response.Room{Room: roomType}, nil
}

// markSelfPlayer 标记自己
func (s *roomService) markSelfPlayer(room *types.Room, code string) {
	if room == nil {
		return
	}
	for _, player := range room.Players {
		player.IsYou = player.Code == code
	}
}

// ensureJoinPlayer 确保玩家加入房间
func (s *roomService) ensureJoinPlayer(ctx context.Context, req *request.JoinRoomRequest, userID uint, code string) (*dbmodels.Player, error) {
	player, err := s.srvCtx.PlayerRepo.GetPlayerByCode(ctx, code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if req.IsVirtualPlayer() {
		if player != nil {
			return nil, errcode.ErrCodePlayerAlreadyJoined
		}
		player = &dbmodels.Player{
			Code:     code,
			RoomID:   req.RoomID,
			NickName: req.PlayerNickName,
			Avatar:   req.PlayerAvatar,
			Type:     req.PlayerType,
			IsOnline: true,
		}
		return player, s.srvCtx.PlayerRepo.CreatePlayer(ctx, player)
	}

	// 如果是真实玩家，允许重复加入房间
	if player != nil {
		player.IsOnline = true
		return player, s.srvCtx.PlayerRepo.UpdatePlayer(ctx, player.Code, player)
	}

	player = &dbmodels.Player{
		Code:     code,
		RoomID:   req.RoomID,
		NickName: req.PlayerNickName,
		Avatar:   req.PlayerAvatar,
		Type:     req.PlayerType,
		UserID:   ptrx.Uint(userID),
		IsOnline: true,
	}
	return player, s.srvCtx.PlayerRepo.CreatePlayer(ctx, player)
}

func (s *roomService) LeaveRoom(ctx context.Context, req *request.LeaveRoomRequest) (err error) {
	if req == nil {
		return errcode.ErrCodeInvalidRequest
	}

	userClaims, err := middlewares.TokenClaimsFromContext(ctx)
	if err != nil {
		return errcode.ErrCodeNoToken
	}

	defer func() {
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = errcode.ErrCodePlayerNotFound
			}
		}
	}()

	// reallyLeave 为 true 时，删除玩家
	if req.ReallyLeave {
		err = s.srvCtx.PlayerRepo.DeletePlayer(ctx, req.PlayerCode)
		if err != nil {
			return err
		}
	} else {
		// 标记玩家离线
		playerObj, err := s.srvCtx.PlayerRepo.GetPlayerByCode(ctx, req.PlayerCode)
		if err != nil {
			return err
		}

		playerObj.IsOnline = false
		err = s.srvCtx.PlayerRepo.UpdatePlayer(ctx, req.PlayerCode, playerObj)
		if err != nil {
			return err
		}
	}

	// session 退出房间
	roomID := types.SocketRoomID(req.RoomID)
	if err := s.srvCtx.SessionManager.LeaveRoom(ctx, userClaims.UUID, roomID); err != nil {
		return err
	}

	// 发布玩家离开房间事件
	leaveMsg := &constant.LeaveRoomMessage{
		UserID:     userClaims.User.ID,
		RoomID:     req.RoomID,
		PlayerCode: req.PlayerCode,
	}

	if err := s.publishRoomEvent(ctx, constant.EventPlayerLeaveRoom, leaveMsg); err != nil {
		return fmt.Errorf("publish room event failed: %w", err)
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

func (s *roomService) publishRoomEvent(ctx context.Context, event string, data any) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal data failed: %w", err)
	}

	msg := &longnet.MemoryQueueMessage{
		Event: event,
		Data:  bytes,
	}

	return s.srvCtx.EventQueue.Publish(ctx, constant.BilliardMessageChannel, msg)
}
