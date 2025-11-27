package cache

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"
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

func (r *VerifyCodeRepository) generateDigitCode(length int) (string, error) {
	var builder strings.Builder
	for range length {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		builder.WriteByte(byte('0' + n.Int64()))
	}
	return builder.String(), nil
}

// GenerateCode 生成并缓存验证码
func (r *VerifyCodeRepository) GenerateCode(ctx context.Context, email string, ttl time.Duration) (string, error) {
	code, err := r.generateDigitCode(6)
	if err != nil {
		return "", err
	}

	cacheKey := EmailCodeCache(email, Withexpire(ttl))
	err = cacheKey.Set(ctx, r.client, code)
	if err != nil {
		return "", err
	}

	return code, nil
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
