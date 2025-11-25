package request

import (
	"github.com/superwhys/billiard-helper/internal/models/errcode"
	"github.com/superwhys/billiard-helper/internal/models/types"
)

type CreateRoomRequest struct{}

type GetRoomRequest struct {
	RoomID uint `json:"room_id" uri:"room_id" form:"room_id" binding:"required"`
}

type GetUserRoomsRequest struct{}

type JoinRoomRequest struct {
	RoomID     uint             `json:"room_id" binding:"required"`
	PlayerType types.PlayerType `json:"player_type" binding:"required"`
	// PlayerNickName 虚拟用户必填
	PlayerNickName string `json:"player_nick_name"`
}

func (r *JoinRoomRequest) IsVirtualPlayer() bool {
	return r.PlayerType == types.PlayerTypeVirtual
}

func (r *JoinRoomRequest) Validate() error {
	if r.IsVirtualPlayer() {
		if r.PlayerNickName == "" {
			return errcode.ErrCodeInvalidRequest
		}
	}
	return nil
}

type LeaveRoomRequest struct {
	RoomID     uint   `json:"room_id" binding:"required"`
	UserID     uint   `json:"user_id" binding:"required"`
	PlayerCode string `json:"player_code" binding:"required"`
	// 是否真的离开房间，如果为 false，则只是标记玩家离线
	ReallyLeave bool `json:"really_leave" binding:"required"`
}

type DeleteRoomRequest struct {
	RoomID uint `json:"room_id" uri:"room_id" form:"room_id" binding:"required"`
}
