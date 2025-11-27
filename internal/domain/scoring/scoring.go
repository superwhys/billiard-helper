package scoring

type Score struct {
	ID         uint   `json:"id"`
	RoomID     uint   `json:"room_id"`
	PlayerID   uint   `json:"player_id"`
	OperatorID uint   `json:"operator_id"` // 操作人
	Type       string `json:"type"`        // 比如 "add", "foul"
	Value      int    `json:"value"`       // 分值变化
	Context    string `json:"context"`     // 额外信息(JSON字符串)，比如进球详情
}
