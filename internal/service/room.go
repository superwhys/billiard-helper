package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/miebyte/goutils/utils/ptrx"
	"github.com/superwhys/billiard-helper/api/middlewares"
	"github.com/superwhys/billiard-helper/internal/dal/cache"
	"github.com/superwhys/billiard-helper/internal/models/constant"
	"github.com/superwhys/billiard-helper/internal/models/dbmodels"
	"github.com/superwhys/billiard-helper/internal/models/errcode"
	"github.com/superwhys/billiard-helper/internal/models/request"
	"github.com/superwhys/billiard-helper/internal/models/response"
	"github.com/superwhys/billiard-helper/internal/models/types"
	"github.com/superwhys/billiard-helper/internal/pkg/codegen"
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
		UserID: userID,
		Status: types.RoomStatusPending,
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

	if err := req.Validate(); err != nil {
		return nil, err
	}

	userClaims, err := middlewares.TokenClaimsFromContext(ctx)
	if err != nil {
		return nil, errcode.ErrCodeNoToken
	}
	user := userClaims.User

	// 加锁，防止并发操作房间
	rdb := s.srvCtx.RedisClient
	roomLock := cache.RoomLockCache(fmt.Sprintf("room:%d", req.RoomID))
	err = roomLock.Lock(ctx, rdb)
	if err != nil {
		return nil, err
	}
	defer roomLock.Unlock(ctx, rdb)

	if req.PlayerNickName == "" {
		req.PlayerNickName = user.Name
	}

	// 检查房间是否存在
	isExist, err := s.srvCtx.RoomRepo.IsRoomExist(ctx, req.RoomID)
	if err != nil {
		return nil, err
	}
	if !isExist {
		return nil, errcode.ErrCodeRoomNotFound
	}

	var playerObj *dbmodels.Player
	if req.IsVirtualPlayer() {
		playerObj, err = s.joinVirtualPlayer(ctx, req)
	} else {
		playerObj, err = s.joinRealPlayer(ctx, req, user)
	}

	playerType := playerObj.ToType()

	// 获取房间信息以及房间内玩家信息
	room, err := s.srvCtx.RoomRepo.GetRoom(ctx, req.RoomID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrCodeRoomNotFound
		}
		return nil, err
	}
	roomType := room.ToType()
	s.markSelfPlayer(roomType, playerType.Code)

	// 获取用户 session 并加入房间
	roomID := types.SocketRoomID(roomType.ID)
	if err := s.srvCtx.SessionManager.JoinRoom(ctx, user.ID, userClaims.SessionID, roomID); err != nil {
		return nil, err
	}

	// 发布玩家加入房间事件
	joinMsg := &constant.JoinRoomMessage{
		EventMsgBase: constant.EventMsgBase{
			UserID:    user.ID,
			RoomID:    roomID,
			SessionID: userClaims.SessionID,
		},
		Player: playerType,
	}

	if err := s.publishRoomEvent(ctx, constant.EventPlayerJoinRoom, joinMsg); err != nil {
		return nil, fmt.Errorf("publish room event failed: %w", err)
	}

	return &response.Room{Room: roomType}, nil
}

func (s *roomService) joinVirtualPlayer(ctx context.Context, req *request.JoinRoomRequest) (*dbmodels.Player, error) {
	code := codegen.GeneratePlayerCode(req.RoomID, req.PlayerType, req.PlayerNickName)

	playerObj, err := s.srvCtx.PlayerRepo.GetPlayerByCode(ctx, code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if playerObj != nil {
		return nil, errcode.ErrCodePlayerAlreadyJoined
	}

	playerObj = &dbmodels.Player{
		Code:     code,
		RoomID:   req.RoomID,
		NickName: req.PlayerNickName,
		Type:     req.PlayerType,
		IsOnline: true,
	}
	return playerObj, s.srvCtx.PlayerRepo.CreatePlayer(ctx, playerObj)
}

func (s *roomService) joinRealPlayer(ctx context.Context, req *request.JoinRoomRequest, user *types.User) (*dbmodels.Player, error) {
	code := codegen.GeneratePlayerCode(req.RoomID, req.PlayerType, req.PlayerNickName)

	playerObj, err := s.srvCtx.PlayerRepo.GetPlayerByCode(ctx, code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 如果是真实玩家，允许重复加入房间
	if playerObj != nil {
		playerObj.IsOnline = true
		return playerObj, s.srvCtx.PlayerRepo.UpdatePlayer(ctx, code, playerObj)
	}

	playerObj = &dbmodels.Player{
		Code:     code,
		RoomID:   req.RoomID,
		NickName: req.PlayerNickName,
		Type:     req.PlayerType,
		UserID:   ptrx.Uint(user.ID),
		IsOnline: true,
	}
	return playerObj, s.srvCtx.PlayerRepo.CreatePlayer(ctx, playerObj)
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

func (s *roomService) LeaveRoom(ctx context.Context, req *request.LeaveRoomRequest) (err error) {
	if req == nil {
		return errcode.ErrCodeInvalidRequest
	}

	rdb := s.srvCtx.RedisClient
	roomLock := cache.RoomLockCache(fmt.Sprintf("room:%d", req.RoomID))
	err = roomLock.Lock(ctx, rdb)
	if err != nil {
		return err
	}
	defer roomLock.Unlock(ctx, rdb)

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

	// 获取玩家信息
	playerObj, err := s.srvCtx.PlayerRepo.GetPlayerByCode(ctx, req.PlayerCode)
	if err != nil {
		return err
	}

	// 获取房间信息
	roomObj, err := s.srvCtx.RoomRepo.GetRoom(ctx, req.RoomID)
	if err != nil {
		return err
	}

	isRoomOwner := roomObj.UserID == userClaims.User.ID
	// 如果不是房主，则只能退出自己
	if !isRoomOwner && playerObj.UserID != nil {
		if ptrx.UintValue(playerObj.UserID) != userClaims.User.ID {
			return errcode.ErrCodePlayerNotAllowed
		}
	}

	// reallyLeave 为 true 时，删除玩家
	if req.ReallyLeave {
		err = s.srvCtx.PlayerRepo.DeletePlayer(ctx, req.PlayerCode)
		if err != nil {
			return err
		}
	} else {
		// 标记玩家离线
		playerObj.IsOnline = false
		err = s.srvCtx.PlayerRepo.UpdatePlayer(ctx, req.PlayerCode, playerObj)
		if err != nil {
			return err
		}
	}

	// 发布玩家离开房间事件
	leaveMsg := &constant.LeaveRoomMessage{
		EventMsgBase: constant.EventMsgBase{
			UserID:    userClaims.User.ID,
			RoomID:    types.SocketRoomID(roomObj.ID),
			SessionID: userClaims.SessionID,
		},
		PlayerCode:   req.PlayerCode,
		PlayerUserID: playerObj.UserID,
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

	rdb := s.srvCtx.RedisClient
	roomLock := cache.RoomLockCache(fmt.Sprintf("room:%d", req.RoomID))
	err := roomLock.Lock(ctx, rdb)
	if err != nil {
		return err
	}
	defer roomLock.Unlock(ctx, rdb)

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
