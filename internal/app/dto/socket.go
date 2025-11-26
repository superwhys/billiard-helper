package dto

type EventMsgBase struct {
	UserID    uint   `json:"user_id"`
	SessionID string `json:"session_id"`
	RoomID    string `json:"room_id"`
}

type JoinRoomMessage struct {
	EventMsgBase
	Player *Player `json:"player"`
}

type LeaveRoomMessage struct {
	EventMsgBase
	PlayerCode   string `json:"player_code"`
	PlayerUserID *uint  `json:"player_user_id"`
}

type ScoreUpdateMessage struct {
	EventMsgBase
	Score *Score `json:"score"`
}
