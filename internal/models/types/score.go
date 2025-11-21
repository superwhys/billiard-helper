package types

type Scores struct {
	PlayerID   uint `json:"player_id"`
	RoomID     uint `json:"room_id"`
	OperatorID uint `json:"operator_id"`
	Change     int  `json:"change"`
	Total      int  `json:"total"`

	Operator *Player `json:"operator,omitempty"`
}
