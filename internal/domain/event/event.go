package event

type Event struct {
	ID         uint `json:"id"`
	MatchID    uint `json:"match_id"`
	PlayerID   uint `json:"player_id"`
	OperatorID uint `json:"operator_id"` // 操作人

	Context string `json:"context"` // 额外信息(JSON字符串)，比如进球详情
}
