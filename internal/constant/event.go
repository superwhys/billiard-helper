package constant

import "fmt"

const (
	EventClientConnectSuccess = "connect"
	EventPlayerJoinRoom       = "player_join_room"
	EventPlayerLeaveRoom      = "player_leave_room"
	EventMatchStarted         = "match_started"
	EventMatchEnded           = "match_ended"
	EventPlayerScoreAdd       = "player_score_add"
	EventPlayerScoreMinus     = "player_score_minus"
	EventPlayerScoreReset     = "player_score_reset"
	EventPlayerScoreUndo      = "player_score_undo"
)

const (
	BilliardMessageChannel = "billiard_message"
)

func MatchRoomID(matchID uint) string {
	return fmt.Sprintf("match_%d", matchID)
}
