package cache

import (
	"context"
	"fmt"

	"github.com/miebyte/goutils/redisutils"
)

// Locker 包装了具体的锁逻辑，绑定了 Redis 客户端
type Locker struct {
	client   *redisutils.RedisClient
	cacheKey *Cachekey
}

// Lock 尝试获取锁
// 注意：这是一个非阻塞锁，如果获取失败会立即返回错误
func (l *Locker) Lock(ctx context.Context) error {
	return l.cacheKey.Lock(ctx, l.client)
}

// Unlock 释放锁
func (l *Locker) Unlock(ctx context.Context) error {
	return l.cacheKey.Unlock(ctx, l.client)
}

// LockManager 锁管理器
type LockManager struct {
	client *redisutils.RedisClient
}

func NewLockManager(client *redisutils.RedisClient) *LockManager {
	return &LockManager{
		client: client,
	}
}

// MatchLock 获取比赛相关的锁
func (m *LockManager) MatchLock(matchID uint) *Locker {
	ck := MatchLockCache(fmt.Sprintf("%d", matchID))
	return &Locker{
		client:   m.client,
		cacheKey: ck,
	}
}
