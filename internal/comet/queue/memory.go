package queue

import (
	"context"
	"encoding/json"
	"fmt"

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

func (m *MemoryQueue) Publish(ctx context.Context, channel string, data *QueueMessage) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal data failed: %w", err)
	}

	ch, ok := m.queues.Get(channel)
	if !ok {
		return fmt.Errorf("channel %s not found", channel)
	}

	ch <- bytes
	return nil
}
