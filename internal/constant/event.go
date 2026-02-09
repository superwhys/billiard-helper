package constant

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/miebyte/goutils/logging"
)

// EventType 事件类型
type EventType = string

const (
	EventClientConnectSuccess EventType = "connect"
	EventPlayerJoinRoom       EventType = "player_join_room"
	EventPlayerLeaveRoom      EventType = "player_leave_room"
	EventMatchStarted         EventType = "match_started"
	EventMatchEnded           EventType = "match_ended"
	EventPlayerScoreAdd       EventType = "player_score_add"
	EventPlayerScoreMinus     EventType = "player_score_minus"
	EventPlayerScoreReset     EventType = "player_score_reset"
	EventPlayerScoreUndo      EventType = "player_score_undo"
	EventMatchRoundNext       EventType = "match_round_next"
)

const (
	BilliardEventChannel = "match_event"
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
