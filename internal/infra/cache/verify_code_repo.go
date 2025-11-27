package cache

import (
	"context"
	"time"

	"github.com/miebyte/goutils/redisutils"
	"github.com/superwhys/billiard-helper/internal/domain/user"
)

var _ user.IVerifyCodeRepository = (*VerifyCodeRepository)(nil)

type VerifyCodeRepository struct {
	client *redisutils.RedisClient
}

func NewVerifyCodeRepository(client *redisutils.RedisClient) *VerifyCodeRepository {
	return &VerifyCodeRepository{
		client: client,
	}
}

// SetCode 存储验证码 (设置过期时间)
func (r *VerifyCodeRepository) SetCode(ctx context.Context, email string, code string, ttl time.Duration) error {
	cacheKey := EmailCodeCache(email, Withexpire(ttl))
	return cacheKey.Set(ctx, r.client, code)
}

// GetCode 获取验证码
func (r *VerifyCodeRepository) GetCode(ctx context.Context, email string) (string, error) {
	cacheKey := EmailCodeCache(email)
	return cacheKey.Get(ctx, r.client)
}

// DeleteCode 删除验证码 (验证成功后)
func (r *VerifyCodeRepository) DeleteCode(ctx context.Context, email string) error {
	cacheKey := EmailCodeCache(email)
	return cacheKey.Del(ctx, r.client)
}
