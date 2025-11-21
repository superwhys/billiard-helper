package emailcode

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/redisutils"
	"github.com/redis/go-redis/v9"
	"github.com/superwhys/billiard-helper/internal/dal/cache"
	"github.com/superwhys/billiard-helper/internal/models/errcode"
	"github.com/superwhys/billiard-helper/internal/models/types"
)

func GetEmailCode(ctx context.Context, redisClient *redisutils.RedisClient, scene types.EmailCodeScene, email string) (string, error) {
	codeCache := cache.EmailCodeCache(BuildEmailCodeKey(scene, email))
	code, err := codeCache.Get(ctx, redisClient)
	if err != nil && !errors.Is(err, redis.Nil) {
		logging.Errorc(ctx, "get email code failed: %v", err)
		return "", errcode.ErrCodeSendEmailCodeFailed
	}

	if errors.Is(err, redis.Nil) {
		return "", errcode.ErrCodeSendEmailCodeFailed
	}

	return code, nil
}

func DeleteEmailCode(ctx context.Context, redisClient *redisutils.RedisClient, scene types.EmailCodeScene, email string) error {
	codeCache := cache.EmailCodeCache(BuildEmailCodeKey(scene, email))
	return codeCache.Del(ctx, redisClient)
}

func BuildEmailCodeKey(scene types.EmailCodeScene, email string) string {
	return fmt.Sprintf("%d:%s", scene, email)
}

func GenerateDigitCode(length int) (string, error) {
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

func NormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}
