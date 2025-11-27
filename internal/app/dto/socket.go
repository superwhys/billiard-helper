package dto

type EventMsgBase struct {
	UserID    uint   `json:"user_id"`
	SessionID string `json:"session_id"`
	MatchID   uint   `json:"match_id"`
}

type JoinMatchEventMessage struct {
	EventMsgBase
	Player *Player `json:"player"`
}

type LeaveMatchEventMessage struct {
	EventMsgBase
	PlayerCode string `json:"player_code"`
}

type MatchStartedEventMessage struct {
	EventMsgBase
}

type MatchEndedEventMessage struct {
	EventMsgBase
}

type MatchScoreUpdateEventMessage struct {
	EventMsgBase
	Score *Score `json:"score"`
}
