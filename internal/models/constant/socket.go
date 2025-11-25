package constant

import "github.com/superwhys/billiard-helper/internal/models/types"

const (
	BilliardNamespace      = "billiard"
	BilliardMessageChannel = "billiard:message:all"
)

const (
	EventClientConnectSuccess = "connect"
	EventPlayerJoinRoom       = "player_join_room"
	EventPlayerLeaveRoom      = "player_leave_room"
	EventPlayerScoreAdd       = "player_score_add"
	EventPlayerScoreMinus     = "player_score_minus"
	EventPlayerScoreReset     = "player_score_reset"
	EventPlayerKickPlayer     = "player_kick_player"
)

type EventMsgBase struct {
	UserID    uint   `json:"user_id"`
	SessionID string `json:"session_id"`
	RoomID    string `json:"room_id"`
}

type JoinRoomMessage struct {
	EventMsgBase
	Player *types.Player `json:"player"`
}

type LeaveRoomMessage struct {
	EventMsgBase
	PlayerCode   string `json:"player_code"`
	PlayerUserID *uint  `json:"player_user_id"`
}
