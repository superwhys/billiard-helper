package request

type AddScoreRequest struct {
	RoomID     uint `json:"room_id" binding:"required"`
	PlayerID   uint `json:"player_id" binding:"required"`
	OperatorID uint `json:"operator_id" binding:"required"`
	Change     int  `json:"change" binding:"required"`
}

type MinusScoreRequest struct {
	RoomID     uint `json:"room_id" binding:"required"`
	PlayerID   uint `json:"player_id" binding:"required"`
	OperatorID uint `json:"operator_id" binding:"required"`
	Change     int  `json:"change" binding:"required"`
}

type ResetScoreRequest struct {
	RoomID     uint `json:"room_id" binding:"required"`
	PlayerID   uint `json:"player_id" binding:"required"`
	OperatorID uint `json:"operator_id" binding:"required"`
}

type GetRoomScoresRequest struct {
	RoomID uint `json:"room_id" binding:"required"`
}
