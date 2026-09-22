package event

import (
	"encoding/json"

	"github.com/superwhys/billiard-helper/internal/constant"
)

type ScoreAction struct {
	PlayerIds []uint `json:"player_ids"`
	Score     int    `json:"score"`
}

type EventData[C any] struct {
	BeforeScores json.RawMessage `json:"before_scores,omitempty"`
	ScoreActions []ScoreAction   `json:"score_actions"`
	Context      C               `json:"context"`
}

type Event struct {
	ID         uint               `json:"id"`
	MatchID    uint               `json:"match_id"`
	Round      uint               `json:"round"`       // 事件所属比赛轮次
	OperatorID uint               `json:"operator_id"` // 操作人
	EventType  constant.EventType `json:"event_type"`  // 事件类型
	Data       json.RawMessage    `json:"data"`        // 事件数据
}
