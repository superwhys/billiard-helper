package ratelimit

import (
	"time"

	"github.com/miebyte/goutils/redisutils"
	"github.com/ulule/limiter/v3"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
)

// NewRedisLimiter 创建基于 Redis 的限流器
func NewRedisLimiter(
	max int,
	expiration time.Duration,
	redisClient *redisutils.RedisClient,
	prefix string,
) (*limiter.Limiter, error) {
	if prefix == "" {
		prefix = "limiter"
	}

	store, err := redisstore.NewStoreWithOptions(redisClient, limiter.StoreOptions{
		Prefix:   prefix,
		MaxRetry: 3,
	})
	if err != nil {
		return nil, err
	}

	rate := limiter.Rate{
		Period: expiration,
		Limit:  int64(max),
	}

	return limiter.New(store, rate), nil
}
