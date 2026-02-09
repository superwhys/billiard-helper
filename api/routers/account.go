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

func AccountGroupRouter(userApp *services.UserApp) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/account"),
		ginutils.WithGroupHandlers(
			ginutils.WithHandler(http.MethodPost, "/send-email-code", SendEmailCodeHandler(userApp)),
			ginutils.WithHandler(http.MethodPost, "/register", AccountRegisterHandler(userApp)),
			ginutils.WithHandler(http.MethodPost, "/login", AccountLoginHandler(userApp)),
			ginutils.WithHandler(http.MethodPost, "/refresh", AccountRefreshHandler(userApp)),
		),
		ginutils.WithGroupHandlers(
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(userApp)),
			ginutils.WithHandler(http.MethodGet, "/me", SelfInfoHandler(userApp)),
			ginutils.WithHandler(http.MethodPost, "/me/update", UpdateSelfInfoHandler(userApp)),
			ginutils.WithHandler(http.MethodPost, "/logout", LogoutHandler(userApp)),
		),
	)
}

// SelfInfoHandler 处理获取用户信息
// @Summary 获取用户信息
// @Description 获取用户信息
// @Tags Account
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ginutils.Ret[dto.User]
// @Router /account/me [get]
func SelfInfoHandler(userApp *services.UserApp) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrUnauthorized) {
			return
		}

		user, err := userApp.GetUserInfo(c.Request.Context(), claims.UserID)
		if handleRouterError(c, err, "get user info failed", errcode.ErrCodeUserGetInfoFailed) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(user))
	}
}

// UpdateSelfInfoHandler 处理更新用户信息
// @Summary 更新用户信息
// @Description 更新用户信息
// @Tags Account
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.UpdateSelfInfoReq true "更新用户信息请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /account/me/update [post]
func UpdateSelfInfoHandler(userApp *services.UserApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.UpdateSelfInfoReq) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrUnauthorized) {
			return
		}

		err = userApp.UpdateSelfInfo(c.Request.Context(), claims.UserID, req)
		if handleRouterError(c, err, "auth update self info handler error", errcode.ErrCodeUserUpdateSelfInfoFailed) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// SendEmailCodeHandler 处理发送邮箱验证码
// @Summary 发送邮箱验证码
// @Description 发送邮箱验证码
// @Tags Account
// @Accept json
// @Produce json
// @Param request body dto.SendRegisterCodeReq true "发送邮箱验证码请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /account/send-email-code [post]
func SendEmailCodeHandler(userApp *services.UserApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.SendRegisterCodeReq) {
		codeId, err := userApp.SendRegisterCode(c.Request.Context(), req)
		if handleRouterError(c, err, "auth send email code handler error", errcode.ErrCodeSendEmailCodeFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseWithData(codeId))
	})
}

// AccountRegisterHandler 处理注册
// @Summary 注册
// @Description 注册
// @Tags Account
// @Accept json
// @Produce json
// @Param request body dto.RegisterReq true "注册请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /account/register [post]
func AccountRegisterHandler(userApp *services.UserApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.RegisterReq) {
		err := userApp.Register(c.Request.Context(), req)
		if handleRouterError(c, err, "auth register handler error", errcode.ErrCodeUserRegisterFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// AccountLoginHandler 处理登录
// @Summary 登录
// @Description 登录
// @Tags Account
// @Accept json
// @Produce json
// @Param request body dto.LoginReq true "登录请求体"
// @Success 200 {object} ginutils.Ret[dto.TokenResponse]
// @Router /account/login [post]
func AccountLoginHandler(userApp *services.UserApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.LoginReq) {
		token, err := userApp.Login(c.Request.Context(), req)
		if handleRouterError(c, err, "auth login handler error", errcode.ErrCodeUserLoginFailed) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(token))
	})
}

// LogoutHandler 处理登出
// @Summary 登出
// @Description 登出
// @Tags Account
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} ginutils.Ret[any]
// @Router /account/logout [post]
func LogoutHandler(userApp *services.UserApp) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := userApp.Logout(c.Request.Context())
		if handleRouterError(c, err, "auth logout handler error", errcode.ErrCodeUserLogoutFailed) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	}
}

// AccountRefreshHandler 刷新 access token
// @Summary 刷新 access token
// @Description 使用 refresh token 刷新 access token
// @Tags Account
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenReq true "刷新 Token 请求体"
// @Success 200 {object} ginutils.Ret[dto.TokenResponse]
// @Router /account/refresh [post]
func AccountRefreshHandler(userApp *services.UserApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.RefreshTokenReq) {
		token, err := userApp.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
		if handleRouterError(c, err, "auth refresh handler error", errcode.ErrInvalidToken) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(token))
	})
}
