package consume

import (
	"context"
	"encoding/json"

	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/utils/ptrx"
	"github.com/superwhys/billiard-helper/internal/models/constant"
)

func (h *Handlers) handlePlayerLeaveRoom(ctx context.Context, data []byte) {
	var msg constant.LeaveRoomMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logging.Errorc(ctx, "unmarshal message failed: %v", err)
		return
	}

	sessions := h.sessionManager.GetSessionsByUserID(msg.UserID)
	if len(sessions) == 0 {
		logging.Errorc(ctx, "user(%d) session not found", msg.UserID)
		return
	}

	// 先广播玩家离开的消息。
	for _, session := range sessions {
		err := h.broadcastRoom(session, msg.RoomID, constant.EventPlayerLeaveRoom, msg.PlayerCode)
		if err != nil {
			logging.Errorc(ctx, "broadcast room failed: %v", err)
		}
	}

	// 移除离开房间玩家的 session
	if msg.PlayerUserID != nil {
		sessions := h.sessionManager.GetSessionsByUserID(ptrx.UintValue(msg.PlayerUserID))
		if len(sessions) == 0 {
			logging.Errorc(ctx, "user(%d) session not found", ptrx.UintValue(msg.PlayerUserID))
			return
		}
		for _, session := range sessions {
			_ = session.Leave(msg.RoomID)
		}
	}
}
