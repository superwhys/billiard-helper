package middlewares

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/redisutils"
	"github.com/superwhys/billiard-helper/api/response"
	"github.com/superwhys/billiard-helper/internal/errcode"
	"github.com/ulule/limiter/v3"

	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
)

const (
	RateLimitExcludeKey = "rate_limit_excluded"
)

func keyGenerator(rateLimitPaths []string) func(c *gin.Context) string {
	return func(c *gin.Context) string {
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		token := c.GetHeader("Authorization")

		for _, p := range rateLimitPaths {
			if strings.HasPrefix(path, p) {

				hash := md5.New()
				hash.Write([]byte(token))
				hash.Write([]byte(path))
				hash.Write([]byte(clientIP))
				key := fmt.Sprintf("%x", hash.Sum(nil))

				logging.Infoc(c.Request.Context(), "limit path: %s clientIp: %s key: %s", path, clientIP, key)
				return key
			}
		}
		return RateLimitExcludeKey
	}
}

func onError(c *gin.Context, err error) {
	ret := response.ErrorResponseWithCode(errcode.ErrSysInternal)
	c.JSON(http.StatusInternalServerError, ret)
}

func onLimitReached(c *gin.Context) {
	ret := response.ErrorResponseWithCode(errcode.ErrTooManyRequests)
	c.JSON(http.StatusTooManyRequests, ret)
}

func RateLimitMiddleware(
	max int,
	expiration time.Duration,
	redisClient *redisutils.RedisClient,
	rateLimitPaths []string,
) gin.HandlerFunc {
	store, err := redisstore.NewStoreWithOptions(redisClient, limiter.StoreOptions{
		Prefix:   "limiter",
		MaxRetry: 3,
	})
	logging.PanicError(err)

	rate := limiter.Rate{
		Period: expiration,
		Limit:  int64(max),
	}

	instance := limiter.New(store, rate)
	return ginlimiter.NewMiddleware(
		instance,
		ginlimiter.WithKeyGetter(keyGenerator(rateLimitPaths)),
		ginlimiter.WithErrorHandler(onError),
		ginlimiter.WithLimitReachedHandler(onLimitReached),
		ginlimiter.WithExcludedKey(func(s string) bool {
			if s == RateLimitExcludeKey {
				return true
			}
			return false
		}),
	)
}
