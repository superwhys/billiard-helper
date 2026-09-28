package handler

import (
	"context"
	"encoding/json"

	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/constant"
)

func (h *Handlers) handlePlayerJoinRoom(ctx context.Context, data []byte) {
	var msg dto.JoinMatchEventMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logging.Errorc(ctx, "unmarshal message failed: %v", err)
		return
	}

	// 广播玩家加入消息
	roomID := constant.SocketRoomID(msg.MatchID)
	err := h.broadcastRoom(ctx, roomID, constant.EventPlayerJoinRoom, msg.Player)
	if err != nil {
		logging.Errorc(ctx, "broadcast room failed: %v", err)
		return
	}

	// 获取该玩家的 socket 连接并加入房间
	sessions := h.socketManager.GetUserSessions(ctx, msg.UserID)
	if len(sessions) == 0 {
		logging.Errorc(ctx, "session not found")
		return
	}

	for _, session := range sessions {
		_ = session.Join(roomID)
	}
}
