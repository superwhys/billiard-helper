package dbmodels

import (
	"gorm.io/gorm"
)

type Room struct {
	gorm.Model
	RoomCode string `gorm:"column:room_code;type:varchar(255);unique;not null;comment:房间代码" json:"room_code"`
	UserID   uint   `gorm:"column:user_id;index;not null;comment:房主ID" json:"user_id"`

	Players []*Player `json:"players"`
	Scores  []*Scores `json:"scores"`
}

func (r *Room) TableName() string {
	return "rooms"
}
