package handler

import (
	"context"
	"encoding/json"

	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/constant"
	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/infra/socket"
)

type EventHandler func(ctx context.Context, data []byte)

type Handlers struct {
	matchService  match.IMatchService
	socketManager *socket.SocketManager
	handlers      map[string]EventHandler
}

func NewHandlers(socketManager *socket.SocketManager) *Handlers {
	h := &Handlers{
		socketManager: socketManager,
		handlers:      make(map[string]EventHandler),
	}
	h.register()
	return h
}

func (h *Handlers) register() {
	h.handlers[constant.EventPlayerJoinRoom] = h.handlePlayerJoinRoom
	h.handlers[constant.EventPlayerLeaveRoom] = h.handlePlayerLeaveRoom
	h.handlers[constant.EventPlayerScoreAdd] = h.handlePlayerScoreAdd
	h.handlers[constant.EventPlayerScoreMinus] = h.handlePlayerScoreMinus
	h.handlers[constant.EventPlayerScoreReset] = h.handlePlayerScoreReset
	h.handlers[constant.EventPlayerScoreUndo] = h.handlePlayerScoreUndo
}

func (h *Handlers) Call(ctx context.Context, event string, data []byte) {
	handler, ok := h.handlers[event]
	if !ok {
		logging.Errorc(ctx, "handler for event %s not found", event)
		return
	}
	handler(ctx, data)
}

func (h *Handlers) broadcastRoom(ctx context.Context, roomID string, event string, data any) error {
	return h.socketManager.BroadcastToRoom(ctx, roomID, event, data)
}

// broadcastScoreEvent 是一个通用的分数事件广播方法，因为分数事件的消息结构都是一样的
func (h *Handlers) broadcastScoreEvent(ctx context.Context, data []byte, event string) {
	var msg dto.MatchScoreUpdateEventMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logging.Errorc(ctx, "unmarshal message failed: %v", err)
		return
	}

	roomID := constant.SocketRoomID(msg.MatchID)
	err := h.broadcastRoom(ctx, roomID, event, msg.Score)
	if err != nil {
		logging.Errorc(ctx, "broadcast room failed: %v", err)
	}
}
