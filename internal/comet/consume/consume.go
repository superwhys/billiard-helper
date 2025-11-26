package consume

import (
	"context"
	"encoding/json"

	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/billiard-helper/internal/comet/manager"
	"github.com/superwhys/billiard-helper/internal/models/constant"
)

type EventHandler func(ctx context.Context, data []byte)

type Handlers struct {
	sessionManager *manager.SessionManager
	handlers       map[string]EventHandler
}

func NewHandlers(sessionManager *manager.SessionManager) *Handlers {
	h := &Handlers{
		sessionManager: sessionManager,
		handlers:       make(map[string]EventHandler),
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
	return h.sessionManager.BroadcastToRoom(ctx, roomID, event, data)
}

// broadcastScoreEvent 是一个通用的分数事件广播方法，因为分数事件的消息结构都是一样的
func (h *Handlers) broadcastScoreEvent(ctx context.Context, data []byte, event string) {
	var msg constant.ScoreUpdateMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logging.Errorc(ctx, "unmarshal message failed: %v", err)
		return
	}

	err := h.broadcastRoom(ctx, msg.RoomID, event, msg.Score)
	if err != nil {
		logging.Errorc(ctx, "broadcast room failed: %v", err)
	}
}
