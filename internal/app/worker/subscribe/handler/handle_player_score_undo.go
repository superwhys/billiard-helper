package handler

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/constant"
)

func (h *Handlers) handlePlayerScoreUndo(ctx context.Context, data []byte) {
	h.broadcastScoreEvent(ctx, data, constant.EventPlayerScoreUndo)
}
