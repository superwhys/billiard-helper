package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superwhys/billiard-helper/api/response"
	"github.com/superwhys/billiard-helper/internal/app/services"
	"github.com/superwhys/billiard-helper/internal/errcode"
	"github.com/superwhys/billiard-helper/internal/pkg/jwt"
)

func SuperAdminMiddleware(adminApp *services.AdminApp) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.ErrorResponseWithCode(errcode.ErrUnauthorized))
			return
		}
		if !adminApp.IsSuperAdmin(claims.UserID) {
			c.AbortWithStatusJSON(http.StatusForbidden, response.ErrorResponseWithCode(errcode.ErrForbidden))
		}
	}
}
