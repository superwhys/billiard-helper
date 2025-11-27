package match

import "context"

type IMatchService interface {
	// CreateRoom 负责创建房间的业务流程
	CreateRoom(ctx context.Context, ownerID uint, config GameConfig) (*Room, error)
	// JoinRoom 处理加入房间，包括各种校验
	JoinRoom(ctx context.Context, roomID uint, userID uint, nickName string) (*Room, error)
	// StartRoom 开始房间
	StartRoom(ctx context.Context, roomID uint) error
	// EndRoom 结束房间
	EndRoom(ctx context.Context, roomID uint) error
	// LeaveRoom 离开房间
	LeaveRoom(ctx context.Context, roomID uint, userID uint) error
	// KickPlayer 踢出玩家
	KickPlayer(ctx context.Context, roomID uint, userID uint) error
}
