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

type JoinRoomMessage struct {
	UserID uint          `json:"user_id`
	RoomID uint          `json:"room_id"`
	Player *types.Player `json:"player"`
}
