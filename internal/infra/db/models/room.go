package models

import "gorm.io/gorm"

type Room struct {
	gorm.Model
	UserID uint  `gorm:"column:user_id;index;not null;comment:房主ID" json:"user_id"`
	Status uint8 `gorm:"column:status;type:tinyint(1);default:1;not null;comment:房间状态" json:"status"`

	Players []*Player `json:"players"`
	Scores  []*Score  `json:"scores"`
}

func (r *Room) TableName() string {
	return "rooms"
}
