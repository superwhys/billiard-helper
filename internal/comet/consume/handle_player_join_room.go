package consume

import (
	"context"
	"encoding/json"

	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/billiard-helper/internal/models/constant"
)

func (h *Handlers) handlePlayerJoinRoom(ctx context.Context, data []byte) {
	var msg constant.JoinRoomMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logging.Errorc(ctx, "unmarshal message failed: %v", err)
		return
	}

	sessions := h.sessionManager.GetSessionsByUserID(msg.UserID)
	if len(sessions) == 0 {
		logging.Errorc(ctx, "user(%d) session not found", msg.UserID)
		return
	}

	for _, session := range sessions {
		err := h.broadcastRoom(session, msg.RoomID, constant.EventPlayerJoinRoom, msg.Player)
		if err != nil {
			logging.Errorc(ctx, "broadcast room failed: %v", err)
		}
	}
}
