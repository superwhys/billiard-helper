package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type MatchEvent struct {
	gorm.Model
	MatchID    uint   `gorm:"column:match_id;index;not null;comment:时间所属比赛 ID" json:"match_id"`
	Round      uint   `gorm:"column:round;index;not null;comment:事件所属比赛轮次" json:"round"`
	OperatorID uint   `gorm:"column:operator_id;index;comment:事件操作人 ID" json:"operator_id"`
	EventType  string `gorm:"column:event_type;index;type:varchar(255);not null;comment:事件类型" json:"event_type"`
	// 事件数据, *event.EventData 类型
	// 但是这里直接使用 JSON 类型
	// 因为 event.EventData.Context 是 map[string]any 类型
	// 这里有不同的玩法自己解析到自己的结构体中
	Data datatypes.JSON `gorm:"column:data;type:json;comment:事件数据" json:"data"`
}

func (e *MatchEvent) TableName() string {
	return "match_events"
}
