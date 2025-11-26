package subscribe

import (
	"context"
	"encoding/json"

	"github.com/miebyte/goutils/logging"
	"github.com/sourcegraph/conc/pool"
	"github.com/superwhys/billiard-helper/internal/app/worker/subscribe/handler"
	"github.com/superwhys/billiard-helper/internal/constant"
	"github.com/superwhys/billiard-helper/internal/domain/shared"
	"github.com/superwhys/billiard-helper/internal/infra/socket"
)

type Subscriber struct {
	eventBus      shared.EventBus
	socketManager *socket.SocketManager
	handlers      *handler.Handlers
}

func NewSubscriber(eventBus shared.EventBus, socketManager *socket.SocketManager) *Subscriber {
	return &Subscriber{
		eventBus:      eventBus,
		socketManager: socketManager,
		handlers:      handler.NewHandlers(socketManager),
	}
}

func (s *Subscriber) Subscribe(ctx context.Context) error {
	ch := s.eventBus.Subscribe(ctx, constant.BilliardMessageChannel)

	worker := pool.New().WithMaxGoroutines(10)
	defer worker.Wait()

	for {
		select {
		case <-ctx.Done():
			return nil
		case data := <-ch:
			worker.Go(func() {
				s.call(ctx, data)
			})
		}
	}
}

func (s *Subscriber) call(ctx context.Context, data []byte) {
	var msg shared.QueueMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logging.Errorc(ctx, "unmarshal message failed: %v", err)
		return
	}

	logging.Debugc(ctx, "received event(%s): %s", msg.Event, string(msg.Data))

	s.handlers.Call(ctx, msg.Event, msg.Data)
}
