package handler

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/constant"
)

func (h *Handlers) handlePlayerScoreReset(ctx context.Context, data []byte) {
	h.broadcastScoreEvent(ctx, data, constant.EventPlayerScoreReset)
}
