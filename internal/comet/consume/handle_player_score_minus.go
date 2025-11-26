package consume

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/models/constant"
)

func (h *Handlers) handlePlayerScoreMinus(ctx context.Context, data []byte) {
	h.broadcastScoreEvent(ctx, data, constant.EventPlayerScoreMinus)
}
