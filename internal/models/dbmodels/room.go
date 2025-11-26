package dbmodels

import (
	"fmt"

	"github.com/superwhys/billiard-helper/internal/models/types"
	"gorm.io/gorm"
)

type Room struct {
	gorm.Model
	UserID uint             `gorm:"column:user_id;index;not null;comment:房主ID" json:"user_id"`
	Status types.RoomStatus `gorm:"column:status;type:tinyint(1);default:1;not null;comment:房间状态" json:"status"`

	Players []*Player `json:"players"`
	Scores  []*Scores `json:"scores"`
}

func (r *Room) TableName() string {
	return "rooms"
}

func (r *Room) ToType() *types.Room {
	roomT := &types.Room{
		ID:     r.ID,
		UserID: r.UserID,
		Status: r.Status,
	}

	if len(r.Players) > 0 {
		roomT.Players = make([]*types.Player, 0, len(r.Players))
		for _, p := range r.Players {
			roomT.Players = append(roomT.Players, p.ToType())
		}
	}

	if len(r.Scores) > 0 {
		roomT.Scores = make([]*types.Scores, 0, len(r.Scores))
		for _, s := range r.Scores {
			roomT.Scores = append(roomT.Scores, s.ToType())
		}
	}

	return roomT
}

func (r *Room) SocketRoomID() string {
	return fmt.Sprintf("room_%d", r.ID)
}
