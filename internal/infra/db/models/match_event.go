package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type MatchEvent struct {
	gorm.Model
	MatchID    uint              `gorm:"column:match_id;index;not null;comment:时间所属比赛 ID" json:"match_id"`
	PlayerID   uint              `gorm:"column:player_id;index;comment:事件涉及的玩家 ID" json:"player_id"`
	OperatorID uint              `gorm:"column:operator_id;index;comment:事件操作人 ID" json:"operator_id"`
	Data       datatypes.JSONMap `gorm:"column:data;type:json;comment:事件数据" json:"data"`

	Operator *Player `gorm:"foreignKey:OperatorID" json:"operator"`
}

func (e *MatchEvent) TableName() string {
	return "match_events"
}
