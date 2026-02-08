package middlewares

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/billiard-helper/api/response"
	"github.com/superwhys/billiard-helper/internal/app/services"
	"github.com/superwhys/billiard-helper/internal/errcode"
	"github.com/superwhys/billiard-helper/internal/pkg/jwt"
)

func TokenVerifyMiddleware(userApp *services.UserApp) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {

			ctx.JSON(http.StatusUnauthorized, response.ErrorResponseWithCode(errcode.ErrNoToken))
			ctx.Abort()
			return
		}

		// Support "Bearer <token>" format
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		tokenStr = strings.TrimSpace(tokenStr)

		claims, err := userApp.GetUserTokenClaims(ctx.Request.Context(), tokenStr)
		if err != nil {
			logging.Errorc(ctx, "verify token failed: %v", err)

			// Handle token expiration specifically
			if errors.Is(err, jwt.ErrTokenExpired) {
				ctx.JSON(http.StatusUnauthorized, response.ErrorResponseWithCode(errcode.ErrTokenExpired))
			} else if ec, ok := errcode.AsErrcode(err); ok {
				ctx.JSON(http.StatusUnauthorized, response.ErrorResponseWithCode(ec))
			} else {
				ctx.JSON(http.StatusUnauthorized, response.ErrorResponseWithCode(errcode.ErrInvalidToken))
			}
			ctx.Abort()
			return
		}

		logging.Debugc(ctx, "token claims: %s", logging.JsonifyNoIndent(claims))

		ctx.Set(string(jwt.TokenContextKey), claims)
		reqCtx := context.WithValue(ctx.Request.Context(), jwt.TokenContextKey, claims)
		ctx.Request = ctx.Request.WithContext(reqCtx)
	}
}
