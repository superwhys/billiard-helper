package dto

import "gorm.io/datatypes"

type MatchEvent struct {
	ID         uint `json:"id"`
	MatchID    uint `json:"match_id"`
	PlayerID   uint `json:"player_id"`
	OperatorID uint `json:"operator_id"`

	EventType string            `json:"event_type"`
	Data      datatypes.JSONMap `json:"data"`
}

type ScoreAction struct {
	PlayerIds []uint `json:"player_ids"`
	Score     uint   `json:"score"`
}

type MatchScoreSyncEvent struct {
	Operator
	MatchID uint `json:"match_id"`
	Round   uint `json:"round"`

	// 用于记录一次操作的所有分数变化
	// 比如一次操作中，A 玩家加分，B,C 玩家扣分
	// 也有可能一次操作只有一个玩家有分数变化，此时 ScoreActions 只有一个元素即可
	ScoreActions []ScoreAction  `json:"score_actions"`
	Context      map[string]any `json:"context"`
}

// MatchScoreUndoReq 撤回分数请求
// 只会撤回指定轮的最新的分数事件
type MatchScoreUndoReq struct {
	Operator
	MatchID uint `json:"match_id"`
	Round   uint `json:"round"`
}

type MatchScoreListReq struct {
	Operator
	MatchID uint `json:"match_id"`
	Round   uint `json:"round"`
}
