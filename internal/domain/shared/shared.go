package shared

import "context"

type EventBus interface {
	Subscribe(ctx context.Context, channel string) <-chan []byte
	Publish(ctx context.Context, channel string, data *QueueMessage) error
}

type QueueMessage struct {
	Event string `json:"event"`
	Data  []byte `json:"data"`
}
