package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/billiard-helper/api/middlewares"
	"github.com/superwhys/billiard-helper/api/response"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/app/services"
	"github.com/superwhys/billiard-helper/internal/errcode"
	"github.com/superwhys/billiard-helper/internal/pkg/jwt"
)

func AdminGroupRouter(adminApp *services.AdminApp) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/admin"),
		ginutils.WithMiddleware(middlewares.SuperAdminMiddleware(adminApp)),
		ginutils.WithHandler(http.MethodGet, "/users", AdminUsersHandler(adminApp)),
		ginutils.WithHandler(http.MethodPost, "/users/block", AdminBlockUserHandler(adminApp)),
		ginutils.WithHandler(http.MethodGet, "/matches/stats", AdminMatchStatsHandler(adminApp)),
	)
}

// AdminUsersHandler 获取用户列表
// @Summary 超管用户列表
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页条数，默认 20，最大 100"
// @Success 200 {object} ginutils.Ret[dto.AdminUsersResponse]
// @Router /admin/users [get]
func AdminUsersHandler(adminApp *services.AdminApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.AdminUsersRequest) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrUnauthorized) {
			return
		}
		result, err := adminApp.ListUsers(c.Request.Context(), claims.UserID, req)
		if handleRouterError(c, err, "admin list users failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(result))
	})
}

// AdminBlockUserHandler 将用户加入黑名单
// @Summary 超管拉黑用户
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.BlockUserRequest true "目标用户"
// @Success 200 {object} ginutils.Ret[dto.BlockUserResponse]
// @Router /admin/users/block [post]
func AdminBlockUserHandler(adminApp *services.AdminApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.BlockUserRequest) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrUnauthorized) {
			return
		}
		result, err := adminApp.BlockUser(c.Request.Context(), claims.UserID, req.UserID)
		if handleRouterError(c, err, "admin block user failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(result))
	})
}

// AdminMatchStatsHandler 获取今日比赛统计
// @Summary 超管今日比赛统计
// @Description 按北京时间统计今日创建的全部比赛，包含已软删除的比赛
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ginutils.Ret[dto.AdminMatchStats]
// @Router /admin/matches/stats [get]
func AdminMatchStatsHandler(adminApp *services.AdminApp) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrUnauthorized) {
			return
		}
		result, err := adminApp.TodayMatchStats(c.Request.Context(), claims.UserID)
		if handleRouterError(c, err, "admin match statistics failed", errcode.ErrSysInternal) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(result))
	}
}
