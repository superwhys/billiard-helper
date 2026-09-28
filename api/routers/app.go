package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/billiard-helper/api/middlewares"
	"github.com/superwhys/billiard-helper/api/response"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/app/services"
	"github.com/superwhys/billiard-helper/internal/pkg/jwt"
)

func AppGroupRouter(userApp *services.UserApp, adminApp *services.AdminApp) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/app"),
		ginutils.WithMiddleware(middlewares.OptionalTokenVerifyMiddleware(userApp)),
		ginutils.WithHandler(http.MethodGet, "/state", AppStateHandler(adminApp)),
	)
}

// AppStateHandler 获取应用状态
// @Summary 获取应用状态
// @Description 未登录时返回默认状态，携带 Token 时返回当前用户的权限状态
// @Tags App
// @Produce json
// @Success 200 {object} ginutils.Ret[dto.AppState]
// @Router /app/state [get]
func AppStateHandler(adminApp *services.AdminApp) gin.HandlerFunc {
	return func(c *gin.Context) {
		state := dto.AppState{}
		if claims, err := jwt.TokenClaimsFromContext(c.Request.Context()); err == nil {
			state.IsSuperAdmin = adminApp.IsSuperAdmin(claims.UserID)
		}
		c.JSON(http.StatusOK, response.ResponseWithData(state))
	}
}
