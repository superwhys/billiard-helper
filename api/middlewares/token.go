package middlewares

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/billiard-helper/api/response"
	"github.com/superwhys/billiard-helper/internal/app/services"
	"github.com/superwhys/billiard-helper/internal/errcode"
	"github.com/superwhys/billiard-helper/internal/pkg/jwt"
)

func TokenVerifyMiddleware(userApp *services.UserApp) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenStr := ctx.GetHeader("Authorization")
		if tokenStr == "" {
			ctx.JSON(http.StatusUnauthorized, response.ErrorResponseWithCode(errcode.ErrCodeNoToken))
			ctx.Abort()
			return
		}

		claims, err := userApp.GetUserTokenClaims(ctx.Request.Context(), tokenStr)
		if err != nil {
			logging.Errorc(ctx, "get secret from jwt token failed: %v", err)
			if ec, ok := errcode.AsErrcode(err); ok {
				ctx.JSON(http.StatusUnauthorized, response.ErrorResponseWithCode(ec))
			} else {
				ctx.JSON(http.StatusUnauthorized, response.ErrorResponseWithCode(errcode.ErrCodeNoToken))
			}
			ctx.Abort()
			return
		}

		ctx.Set(string(jwt.TokenContextKey), claims)

		logging.Debugc(ctx, "token claims: %s", logging.JsonifyNoIndent(claims))
		// 将 claims 注入 request context，供下游以 context.Value 读取
		reqCtx := context.WithValue(ctx.Request.Context(), jwt.TokenContextKey, claims)
		ctx.Request = ctx.Request.WithContext(reqCtx)
	}
}
