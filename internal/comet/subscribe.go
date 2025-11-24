package comet

import (
	"context"
	"encoding/json"

	"github.com/miebyte/goutils/logging"
	"github.com/sourcegraph/conc/pool"
	"github.com/superwhys/billiard-helper/internal/comet/consume"
	"github.com/superwhys/billiard-helper/internal/models/constant"
	"github.com/superwhys/billiard-helper/internal/pkg/longnet"
	"github.com/superwhys/billiard-helper/internal/service"
)

type Subscriber struct {
	srv      *service.Service
	handlers *consume.Handlers
	queue    longnet.EventQueue
}

func NewSubscriber(queue longnet.EventQueue, sessionManager longnet.ISessionManager, srv *service.Service) *Subscriber {
	return &Subscriber{
		queue:    queue,
		srv:      srv,
		handlers: consume.NewHandlers(srv, sessionManager),
	}
}

func (c *Subscriber) Subscribe(ctx context.Context) error {
	ch := c.queue.Subscribe(ctx, constant.BilliardMessageChannel)

	worker := pool.New().WithMaxGoroutines(10)
	defer worker.Wait()

	for {
		select {
		case <-ctx.Done():
			return nil
		case data := <-ch:
			worker.Go(func() {
				c.call(ctx, data)
			})
		}
	}
}

func (c *Subscriber) call(ctx context.Context, data []byte) {
	var msg longnet.MemoryQueueMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		logging.Errorc(ctx, "unmarshal message failed: %v", err)
		return
	}

	logging.Debugc(ctx, "received data: %v", msg)

	c.handlers.Call(ctx, msg.Event, msg.Data)
}
