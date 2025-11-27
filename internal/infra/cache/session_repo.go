package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/miebyte/goutils/redisutils"
	"github.com/superwhys/billiard-helper/internal/domain/user"
)

var _ user.ISessionRepository = (*SessionRepository)(nil)

type SessionRepository struct {
	client *redisutils.RedisClient
}

func NewSessionRepository(client *redisutils.RedisClient) *SessionRepository {
	return &SessionRepository{
		client: client,
	}
}

func (r *SessionRepository) SetSession(ctx context.Context, userID uint, token string, ttl time.Duration) error {
	key := strconv.FormatUint(uint64(userID), 10)
	cacheKey := AuthSessionCache(key, Withexpire(ttl))
	return cacheKey.Set(ctx, r.client, token)
}

func (r *SessionRepository) GetSession(ctx context.Context, userID uint) (string, error) {
	key := strconv.FormatUint(uint64(userID), 10)
	cacheKey := AuthSessionCache(key)
	return cacheKey.Get(ctx, r.client)
}

func (r *SessionRepository) DeleteSession(ctx context.Context, userID uint) error {
	key := strconv.FormatUint(uint64(userID), 10)
	cacheKey := AuthSessionCache(key)
	return cacheKey.Del(ctx, r.client)
}
