package event

import (
	"github.com/superwhys/billiard-helper/internal/constant"
	"gorm.io/datatypes"
)

type Event struct {
	ID         uint               `json:"id"`
	MatchID    uint               `json:"match_id"`
	Round      uint               `json:"round"`       // 事件所属比赛轮次
	OperatorID uint               `json:"operator_id"` // 操作人
	EventType  constant.EventType `json:"event_type"`  // 事件类型
	Data       datatypes.JSONMap  `json:"data"`        // 事件数据
}
