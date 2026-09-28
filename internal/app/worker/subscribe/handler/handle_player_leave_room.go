package handler

import (
	"context"
	"encoding/json"

	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/utils/ptrx"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/constant"
)

func (h *Handlers) handlePlayerLeaveRoom(ctx context.Context, data []byte) {
	var msg dto.LeaveMatchEventMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logging.Errorc(ctx, "unmarshal message failed: %v", err)
		return
	}

	roomID := constant.SocketRoomID(msg.MatchID)
	err := h.broadcastRoom(ctx, roomID, constant.EventPlayerLeaveRoom, msg.PlayerCode)
	if err != nil {
		logging.Errorc(ctx, "broadcast room failed: %v", err)
	}

	playerRepo := h.repoFactory.PlayerRepo()
	player, err := playerRepo.FindByCode(ctx, msg.PlayerCode)
	if err != nil {
		logging.Errorc(ctx, "find player by code failed: %v", err)
		return
	}

	if player.UserID != nil {
		// 获取该玩家的 session 并逐一离开房间
		sessions := h.socketManager.GetUserSessions(ctx, ptrx.UintValue(player.UserID))
		if len(sessions) == 0 {
			logging.Errorc(ctx, "session not found")
			return
		}
		for _, session := range sessions {
			_ = session.Leave(roomID)
		}
	}
}
