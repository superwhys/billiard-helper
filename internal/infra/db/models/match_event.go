package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type MatchEvent struct {
	gorm.Model
	MatchID    uint              `gorm:"column:match_id;index;not null;comment:时间所属比赛 ID" json:"match_id"`
	Round      uint              `gorm:"column:round;index;not null;comment:事件所属比赛轮次" json:"round"`
	OperatorID uint              `gorm:"column:operator_id;index;comment:事件操作人 ID" json:"operator_id"`
	EventType  string            `gorm:"column:event_type;index;type:varchar(255);not null;comment:事件类型" json:"event_type"`
	Data       datatypes.JSONMap `gorm:"column:data;type:json;comment:事件数据" json:"data"`

	Operator *Player `gorm:"foreignKey:OperatorID" json:"operator"`
}

func (e *MatchEvent) TableName() string {
	return "match_events"
}
