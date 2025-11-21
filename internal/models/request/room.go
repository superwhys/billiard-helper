package request

import "github.com/superwhys/billiard-helper/internal/models/types"

type CreateRoomRequest struct {
	RoomCode string `json:"room_code" binding:"required"`
}

type GetRoomRequest struct {
	RoomID uint `json:"room_id" uri:"room_id" form:"room_id" binding:"required"`
}

type GetUserRoomsRequest struct{}

type JoinRoomRequest struct {
	RoomID         uint             `json:"room_id" binding:"required"`
	PlayerType     types.PlayerType `json:"player_type"`
	PlayerNickName string           `json:"player_nick_name"`
	PlayerAvatar   string           `json:"player_avatar"`
}

type LeaveRoomRequest struct {
	RoomID     uint   `json:"room_id" binding:"required"`
	PlayerCode string `json:"player_code" binding:"required"`
}

type DeleteRoomRequest struct {
	RoomID uint `json:"room_id" uri:"room_id" form:"room_id" binding:"required"`
}
