package eventbus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/redisutils"
	"github.com/superwhys/billiard-helper/internal/domain/shared"
)

const (
	BilliardMessageChannelPrefix = "billiard"
)

var _ shared.EventBus = (*RedisEventBus)(nil)

// RedisEventBus 基于 Redis Pub/Sub 实现的事件总线
type RedisEventBus struct {
	client *redisutils.RedisClient
}

func NewRedisEventBus(client *redisutils.RedisClient) *RedisEventBus {
	return &RedisEventBus{client: client}
}

func (b *RedisEventBus) channel(channel string) string {
	return fmt.Sprintf("%s:%s", BilliardMessageChannelPrefix, channel)
}

func (b *RedisEventBus) Subscribe(ctx context.Context, channel string) <-chan []byte {
	sub := b.client.Subscribe(ctx, b.channel(channel))
	ch := make(chan []byte, 100)

	go func() {
		defer close(ch)
		defer sub.Close()

		msgCh := sub.Channel()
		for {
			select {
			case <-ctx.Done():
				logging.Infoc(ctx, "eventbus subscribe channel(%s) stopped: %v", channel, ctx.Err())
				return
			case msg, ok := <-msgCh:
				if !ok {
					logging.Infoc(ctx, "eventbus subscribe channel(%s) closed", channel)
					return
				}
				ch <- []byte(msg.Payload)
			}
		}
	}()

	return ch
}

func (b *RedisEventBus) Publish(ctx context.Context, channel string, data *shared.QueueMessage) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal queue message failed: %w", err)
	}

	channel = b.channel(channel)
	if err := b.client.Publish(ctx, channel, payload).Err(); err != nil {
		return fmt.Errorf("publish to channel(%s) failed: %w", channel, err)
	}

	logging.Debugc(ctx, "eventbus publish channel(%s) event(%s)", channel, data.Event)
	return nil
}
