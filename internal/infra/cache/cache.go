package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"strings"
	"time"

	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/redisutils"
	"github.com/redis/go-redis/v9"
)

const (
	globalPrefix = "billiard-helper"
)

type optionFn func(*Cachekey)

func Withexpire(dura time.Duration) optionFn {
	return func(ck *Cachekey) {
		ck.expire = dura
	}
}

type Cachekey struct {
	key    string
	expire time.Duration
}

func (c *Cachekey) ExpireTime() time.Duration {
	return c.expire
}

func (c *Cachekey) String() string {
	return c.key
}

func (c *Cachekey) Key() string {
	return c.key
}

func (c *Cachekey) Scan(ctx context.Context, client *redisutils.RedisClient) iter.Seq2[string, error] {
	pattern := c.key + ":*"
	cursor := uint64(0)
	return func(yield func(string, error) bool) {
		for {
			keys, cur, err := client.Scan(ctx, cursor, pattern, 100).Result()
			if err != nil {
				yield("", fmt.Errorf("scan pattern(%s) failed: %w", pattern, err))
				return
			}

			for _, key := range keys {
				if !yield(key, nil) {
					logging.Infoc(ctx, "scan pattern(%s) finished", pattern)
					return
				}
			}

			cursor = cur
			if cursor == 0 {
				logging.Infoc(ctx, "scan pattern(%s) finished", pattern)
				return
			}
		}
	}
}

func (c *Cachekey) Get(ctx context.Context, client *redisutils.RedisClient) (string, error) {
	return client.Get(ctx, c.key).Result()
}

func (c *Cachekey) Set(ctx context.Context, client *redisutils.RedisClient, value any) error {
	err := client.Set(ctx, c.key, value, c.expire).Err()
	if err != nil {
		return fmt.Errorf("set cache(%s) failed. error: %w", c.key, err)
	}

	return nil
}

func (c *Cachekey) SetNX(ctx context.Context, client *redisutils.RedisClient, value any) (bool, error) {
	isSet, err := client.SetNX(ctx, c.key, value, c.expire).Result()
	if err != nil {
		return false, fmt.Errorf("setNX cache(%s) failed. error: %w", c.key, err)
	}

	return isSet, nil
}

func (c *Cachekey) Exists(ctx context.Context, client *redisutils.RedisClient) (bool, error) {
	exists, err := client.Exists(ctx, c.key).Result()
	if err != nil {
		return false, fmt.Errorf("exists cache(%s) failed. error: %w", c.key, err)
	}

	return exists > 0, nil
}

func (c *Cachekey) Lock(ctx context.Context, client *redisutils.RedisClient) error {
	err := client.TryLock(ctx, c.key, c.expire)
	if err != nil {
		return fmt.Errorf("lock cache(%s) failed. error: %w", c.key, err)
	}

	return nil
}

func (c *Cachekey) Unlock(ctx context.Context, client *redisutils.RedisClient) error {
	err := client.Unlock(ctx, c.key)
	if err != nil {
		return fmt.Errorf("unlock cache(%s) failed. error: %w", c.key, err)
	}

	return nil
}

func (c *Cachekey) GetValue(ctx context.Context, client *redisutils.RedisClient, obj any) error {
	err := client.GetValue(ctx, c.key, obj)
	if err != nil {
		return fmt.Errorf("get cache(%s) failed. error: %w", c.key, err)
	}

	return nil
}

func (c *Cachekey) SetValue(ctx context.Context, client *redisutils.RedisClient, obj any) error {
	err := client.SetValue(ctx, c.key, obj, c.expire)
	if err != nil {
		return fmt.Errorf("set cache(%s) failed. error: %w", c.key, err)
	}

	return nil
}

func (c *Cachekey) ZAdd(ctx context.Context, client *redisutils.RedisClient, members ...redis.Z) error {
	err := client.ZAdd(ctx, c.key, members...).Err()
	if err != nil {
		return fmt.Errorf("zadd cache(%s) failed. error: %w", c.key, err)
	}

	return nil
}

func (c *Cachekey) ZRangeByScore(ctx context.Context, client *redisutils.RedisClient, min, max string, offset, count int64) ([]string, error) {
	zRangeBy := &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}
	results, err := client.ZRangeByScore(ctx, c.key, zRangeBy).Result()
	if err != nil {
		return nil, fmt.Errorf("zrangebyscore cache(%s) failed. error: %w", c.key, err)
	}

	return results, nil
}

func (c *Cachekey) ZRangeByScoreWithScores(ctx context.Context, client *redisutils.RedisClient, min, max string, offset, count int64) ([]redis.Z, error) {
	zRangeBy := &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: offset,
		Count:  count,
	}

	results, err := client.ZRangeByScoreWithScores(ctx, c.key, zRangeBy).Result()
	if err != nil {
		return nil, fmt.Errorf("zrangebyscorewithscores cache(%s) failed. error: %w", c.key, err)
	}

	return results, nil
}

func (c *Cachekey) ZRem(ctx context.Context, client *redisutils.RedisClient, members ...string) error {
	err := client.ZRem(ctx, c.key, members).Err()
	if err != nil {
		return fmt.Errorf("zrem cache(%s) failed. error: %w", c.key, err)
	}

	return nil
}

func (c *Cachekey) Expire(ctx context.Context, client *redisutils.RedisClient, e time.Duration) error {
	err := client.Expire(ctx, c.key, e).Err()
	if err != nil {
		return fmt.Errorf("set expire of cache(%s) failed. error: %w", c.key, err)
	}

	return nil
}

func (c *Cachekey) SAdd(ctx context.Context, client *redisutils.RedisClient, member any) error {
	err := client.SAdd(ctx, c.key, member).Err()
	if err != nil {
		return fmt.Errorf("sadd member of cache(%s) failed. error: %w", c.key, err)
	}

	return nil
}

// convertValueToRedisArg converts a value to a format suitable for Redis storage
func (c *Cachekey) convertValueToRedisArg(value any) (any, error) {
	switch v := value.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		string,
		bool,
		time.Time, time.Duration,
		[]byte:
		return v, nil
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("json marshal failed: %w", err)
		}
		return string(jsonBytes), nil
	}
}

func (c *Cachekey) LPush(ctx context.Context, client *redisutils.RedisClient, value any) error {
	redisValue, err := c.convertValueToRedisArg(value)
	if err != nil {
		return fmt.Errorf("convert value to redis arg failed: %w", err)
	}

	err = client.LPush(ctx, c.key, redisValue).Err()
	if err != nil {
		return fmt.Errorf("lpush cache(%s) failed. error: %w", c.key, err)
	}

	return nil
}

func (c *Cachekey) RPush(ctx context.Context, client *redisutils.RedisClient, value any) error {
	redisValue, err := c.convertValueToRedisArg(value)
	if err != nil {
		return fmt.Errorf("convert value to redis arg failed: %w", err)
	}

	err = client.RPush(ctx, c.key, redisValue).Err()
	if err != nil {
		return fmt.Errorf("rpush cache(%s) failed. error: %w", c.key, err)
	}

	return nil
}

func (c *Cachekey) LPop(ctx context.Context, client *redisutils.RedisClient) (string, error) {
	result, err := client.LPop(ctx, c.key).Result()
	if err != nil {
		return "", fmt.Errorf("lpop cache(%s) failed. error: %w", c.key, err)
	}

	return result, nil
}

func (c *Cachekey) RPop(ctx context.Context, client *redisutils.RedisClient) (string, error) {
	result, err := client.RPop(ctx, c.key).Result()
	if err != nil {
		return "", fmt.Errorf("rpop cache(%s) failed. error: %w", c.key, err)
	}

	return result, nil
}

func (c *Cachekey) LLen(ctx context.Context, client *redisutils.RedisClient) (int64, error) {
	length, err := client.LLen(ctx, c.key).Result()
	if err != nil {
		return 0, fmt.Errorf("llen cache(%s) failed. error: %w", c.key, err)
	}

	return length, nil
}

func (c *Cachekey) LRange(ctx context.Context, client *redisutils.RedisClient, start, stop int64) ([]string, error) {
	result, err := client.LRange(ctx, c.key, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("lrange cache(%s) failed. error: %w", c.key, err)
	}

	return result, nil
}

func (c *Cachekey) LRem(ctx context.Context, client *redisutils.RedisClient, count int64, value any) error {
	err := client.LRem(ctx, c.key, count, value).Err()
	if err != nil {
		return fmt.Errorf("lrem cache(%s) failed. error: %w", c.key, err)
	}

	return nil
}

func (c *Cachekey) Del(ctx context.Context, client *redisutils.RedisClient) error {
	err := client.Del(ctx, c.key).Err()
	if err != nil {
		return fmt.Errorf("del cache(%s) failed. error: %w", c.key, err)
	}

	return nil
}

func (c *Cachekey) HSet(ctx context.Context, client *redisutils.RedisClient, field string, value any) error {
	err := client.HSet(ctx, c.key, field, value).Err()
	if err != nil {
		return fmt.Errorf("hset cache(%s) failed. error: %w", c.key, err)
	}

	return nil
}

func (c *Cachekey) HGet(ctx context.Context, client *redisutils.RedisClient, field string) (string, error) {
	result, err := client.HGet(ctx, c.key, field).Result()
	if err != nil {
		return "", fmt.Errorf("hget cache(%s) failed. error: %w", c.key, err)
	}

	return result, nil
}

func (c *Cachekey) HKeys(ctx context.Context, client *redisutils.RedisClient) ([]string, error) {
	keys, err := client.HKeys(ctx, c.key).Result()
	if err != nil {
		return nil, fmt.Errorf("hkeys cache(%s) failed. error: %w", c.key, err)
	}

	return keys, nil
}

func (c *Cachekey) HDel(ctx context.Context, client *redisutils.RedisClient, fields ...string) error {
	err := client.HDel(ctx, c.key, fields...).Err()
	if err != nil {
		return fmt.Errorf("hdel cache(%s) failed. error: %w", c.key, err)
	}

	return nil
}

type genCacheWithKeyFn func(key string, opts ...optionFn) *Cachekey

func genCacheWithKey(prefix string, expire time.Duration) genCacheWithKeyFn {
	return func(key string, opts ...optionFn) *Cachekey {
		keyItems := []string{globalPrefix, prefix}
		if key != "" {
			keyItems = append(keyItems, key)
		}

		c := &Cachekey{
			key:    strings.Join(keyItems, ":"),
			expire: expire,
		}
		for _, opt := range opts {
			opt(c)
		}

		return c
	}
}
