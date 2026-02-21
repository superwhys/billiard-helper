package cache

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/miebyte/goutils/redisutils"
	"github.com/redis/go-redis/v9"
	"github.com/superwhys/billiard-helper/internal/domain/user"
	"github.com/superwhys/billiard-helper/internal/errcode"
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

func (r *SessionRepository) SetSession(ctx context.Context, sessionID string, userID uint, ttl time.Duration) error {
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}
	cacheKey := AuthSessionCache(sessionID, Withexpire(ttl))
	value := strconv.FormatUint(uint64(userID), 10)
	return cacheKey.Set(ctx, r.client, value)
}

func (r *SessionRepository) GetSession(ctx context.Context, sessionID string) (uint, error) {
	if sessionID == "" {
		return 0, fmt.Errorf("session id is required")
	}
	cacheKey := AuthSessionCache(sessionID)
	value, err := cacheKey.Get(ctx, r.client)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, errcode.ErrTokenExpired
		}
		return 0, err
	}
	userID, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid session user id")
	}
	return uint(userID), nil
}

func (r *SessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}
	cacheKey := AuthSessionCache(sessionID)
	return cacheKey.Del(ctx, r.client)
}
