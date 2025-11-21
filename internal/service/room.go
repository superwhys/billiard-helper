package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/miebyte/goutils/utils/ptrx"
	"github.com/superwhys/billiard-helper/api/middlewares"
	"github.com/superwhys/billiard-helper/internal/models/dbmodels"
	"github.com/superwhys/billiard-helper/internal/models/errcode"
	"github.com/superwhys/billiard-helper/internal/models/request"
	"github.com/superwhys/billiard-helper/internal/models/response"
	"github.com/superwhys/billiard-helper/internal/models/types"
	"github.com/superwhys/billiard-helper/internal/pkg/hash"
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
	exist, err := s.srvCtx.RoomRepo.IsRoomExist(ctx, req.RoomID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, errcode.ErrCodeRoomNotFound
	}

	// 生成玩家幂等 code
	code := hash.GenerateHash(
		fmt.Sprintf("%d", req.RoomID),
		fmt.Sprintf("%d", userID),
		fmt.Sprintf("%d", req.PlayerType),
		fmt.Sprintf("%s", req.PlayerNickName),
	)

	// 检查玩家是否已经加入房间
	player, err := s.srvCtx.PlayerRepo.GetPlayerByCode(ctx, code)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if player != nil {
		return nil, errcode.ErrCodePlayerAlreadyJoined
	}

	// 创建玩家记录
	playerModel := &dbmodels.Player{
		Code:     code,
		RoomID:   req.RoomID,
		NickName: req.PlayerNickName,
		Avatar:   req.PlayerAvatar,
		Type:     req.PlayerType,
	}

	if req.PlayerType == types.PlayerTypeReal {
		playerModel.UserID = ptrx.Uint(userID)
	}

	err = s.srvCtx.PlayerRepo.CreatePlayer(ctx, playerModel)
	if err != nil {
		return nil, err
	}

	room, err := s.srvCtx.RoomRepo.GetRoom(ctx, req.RoomID)
	if err != nil {
		return nil, err
	}

	roomType := room.ToType()
	for _, player := range roomType.Players {
		player.IsYou = player.Code == code
	}

	return &response.Room{Room: roomType}, nil
}

func (s *roomService) LeaveRoom(ctx context.Context, req *request.LeaveRoomRequest) error {
	if req == nil {
		return errcode.ErrCodeInvalidRequest
	}

	err := s.srvCtx.PlayerRepo.DeletePlayer(ctx, req.PlayerCode)
	if err != nil {
		return fmt.Errorf("delete player failed: %w", err)
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
