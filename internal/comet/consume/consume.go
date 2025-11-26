package consume

import (
	"context"

	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/billiard-helper/internal/models/constant"
	"github.com/superwhys/billiard-helper/internal/pkg/longnet"
	"github.com/superwhys/billiard-helper/internal/service"
)

type EventHandler func(ctx context.Context, data []byte)

type Handlers struct {
	srv            *service.Service
	sessionManager longnet.ISessionManager
	handlers       map[string]EventHandler
}

func NewHandlers(srv *service.Service, sessionManager longnet.ISessionManager) *Handlers {
	h := &Handlers{
		srv:            srv,
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
}

func (h *Handlers) Call(ctx context.Context, event string, data []byte) {
	handler, ok := h.handlers[event]
	if !ok {
		logging.Errorc(ctx, "handler for event %s not found", event)
		return
	}
	handler(ctx, data)
}

func (h *Handlers) broadcastRoom(session longnet.ISession, roomID string, event string, data any) error {
	return session.Namespace().To(roomID).EmitExcept(event, data, session)
}
