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

	// 广播玩家加入消息
	err := h.broadcastRoom(ctx, msg.RoomID, constant.EventPlayerJoinRoom, msg.Player)
	if err != nil {
		logging.Errorc(ctx, "broadcast room failed: %v", err)
		return
	}

	// 获取该玩家的 socket 连接并加入房间
	session := h.sessionManager.GetUserSession(msg.UserID, msg.SessionID)
	if session == nil {
		logging.Errorc(ctx, "session not found")
		return
	}

	_ = session.Join(msg.RoomID)
}
