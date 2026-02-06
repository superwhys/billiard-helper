package dto

type Score struct {
	ID         uint   `json:"id"`
	RoomID     uint   `json:"room_id"`
	PlayerID   uint   `json:"player_id"`
	OperatorID uint   `json:"operator_id"` // 操作人
	Type       string `json:"type"`        // 比如 "add", "foul"
	Value      int    `json:"value"`       // 分值变化
	Context    string `json:"context"`     // 额外信息(JSON字符串)，比如进球详情
}

// AddScoreRequest 加分请求
type AddScoreRequest struct {
	RoomID     uint           `json:"room_id"`
	PlayerID   uint           `json:"player_id"`
	OperatorID uint           `json:"operator_id"`
	Context    map[string]any `json:"context"`
}

// UndoRequest 撤回请求
type UndoRequest struct {
	RoomID     uint `json:"room_id"`
	OperatorID uint `json:"operator_id"`
}

// RoomScoreResponse 房间当前分数统计
type RoomScoreResponse struct {
	RoomID      uint         `json:"room_id"`
	TotalScores map[uint]int `json:"total_scores"` // PlayerID -> TotalScore
	History     []Score      `json:"history"`      // 最近的事件流
}
