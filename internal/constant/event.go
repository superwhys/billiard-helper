package constant

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/miebyte/goutils/logging"
)

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

func SocketRoomID(matchID uint) string {
	return fmt.Sprintf("match_%d", matchID)
}

func IsSocketRoomID(roomID string) bool {
	return strings.HasPrefix(roomID, "match_")
}

func ParseSocketRoomID(roomID string) uint {
	if !IsSocketRoomID(roomID) {
		return 0
	}

	mathIDStr := strings.TrimPrefix(roomID, "match_")
	mathID, err := strconv.ParseUint(mathIDStr, 10, 64)
	if err != nil {
		logging.Errorf("parse socket room id failed: %v", err)
		return 0
	}
	return uint(mathID)
}
