package longnet

import (
	"context"
	"encoding/json"

	"github.com/miebyte/goutils/logging"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type MemoryQueue struct {
	queues cmap.ConcurrentMap[string, chan []byte]
}

func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		queues: cmap.New[chan []byte](),
	}
}

func (m *MemoryQueue) Subscribe(ctx context.Context, channel string) <-chan []byte {
	ch, ok := m.queues.Get(channel)
	if !ok {
		ch = make(chan []byte, 100)
		m.queues.Set(channel, ch)
	}

	return ch
}

func (m *MemoryQueue) Publish(ctx context.Context, channel string, data []byte) {
	msg := MemoryQueueMessage{
		Event: channel,
		Data:  data,
	}

	bytes, err := json.Marshal(msg)
	if err != nil {
		logging.Errorc(ctx, "marshal message failed: %v", err)
		return
	}

	ch, ok := m.queues.Get(channel)
	if !ok {
		logging.Errorc(ctx, "channel %s not found", channel)
		return
	}

	ch <- bytes
}
