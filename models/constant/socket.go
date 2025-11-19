package constant

const (
	BilliardNamespace = "billiard"
)

// Server receive event
const (
	EventClientJoinRoom   = "join_room"
	EventClientLeaveRoom  = "leave_room"
	EventClientScoreAdd   = "score_add"
	EventClientScoreMinus = "score_minus"
	EventClientScoreReset = "score_reset"
	EventClientKickPlayer = "kick_player"
)

// Server callback event
const (
	EventCallbackFailed            = "event_callback_failed"
	EventCallbackJoinRoomSuccess   = "join_room_success"
	EventCallbackJoinRoomFailed    = "join_room_failed"
	EventCallbackLeaveRoomSuccess  = "leave_room_success"
	EventCallbackLeaveRoomFailed   = "leave_room_failed"
	EventCallbackScoreAddSuccess   = "score_add_success"
	EventCallbackScoreMinusSuccess = "score_minus_success"
	EventCallbackScoreResetSuccess = "score_reset_success"
	EventCallbackKickPlayerSuccess = "kick_player_success"
	EventCallbackKickPlayerFailed  = "kick_player_failed"
)

// Server broadcast event
const (
	EventBroadcastRoomJoin       = "broadcast_room_joined"
	EventBroadcastRoomLeave      = "broadcast_room_left"
	EventBroadcastRoomScoreAdd   = "broadcast_score_add"
	EventBroadcastRoomScoreMinus = "broadcast_score_minus"
	EventBroadcastRoomScoreReset = "broadcast_score_reset"
	EventBroadcastRoomKickPlayer = "broadcast_kick_player"
)
