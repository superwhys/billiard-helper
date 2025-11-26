package handler

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/constant"
)

func (h *Handlers) handlePlayerScoreAdd(ctx context.Context, data []byte) {
	h.broadcastScoreEvent(ctx, data, constant.EventPlayerScoreAdd)
}
