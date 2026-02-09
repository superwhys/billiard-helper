package cache

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/miebyte/goutils/redisutils"
	"github.com/superwhys/billiard-helper/internal/domain/user"
)

type VerifyCodePayload struct {
	CodeID    string    `json:"code_id"`
	Account   string    `json:"account"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

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
func (r *VerifyCodeRepository) GenerateCode(ctx context.Context, account string, ttl time.Duration) (string, string, error) {
	code, err := r.generateDigitCode(6)
	if err != nil {
		return "", "", err
	}

	codeId := uuid.New().String()
	cacheKey := VerifyCodeCache(codeId, Withexpire(ttl))

	payload := &VerifyCodePayload{
		CodeID:    codeId,
		Account:   account,
		Code:      code,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}
	err = cacheKey.SetValue(ctx, r.client, payload)
	if err != nil {
		return "", "", err
	}

	return codeId, code, nil
}

// GetCode 获取验证码
func (r *VerifyCodeRepository) GetCode(ctx context.Context, codeId, account string) (string, error) {
	cacheKey := VerifyCodeCache(codeId)

	payload := &VerifyCodePayload{}
	err := cacheKey.GetValue(ctx, r.client, payload)
	if err != nil {
		return "", err
	}

	return payload.Code, nil
}

// DeleteCode 删除验证码 (验证成功后)
func (r *VerifyCodeRepository) DeleteCode(ctx context.Context, codeId string) error {
	cacheKey := VerifyCodeCache(codeId)
	return cacheKey.Del(ctx, r.client)
}
