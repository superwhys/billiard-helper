package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/billiard-helper/internal/models/errcode"
	"github.com/superwhys/billiard-helper/internal/models/request"
	"github.com/superwhys/billiard-helper/internal/models/response"
	"github.com/superwhys/billiard-helper/internal/ports"
)

func AuthGroupRouter(authSvc ports.AuthService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/auth"),
		ginutils.WithHandler(http.MethodPost, "/send-email-code", AuthSendEmailCodeHandler(authSvc)),
		ginutils.WithHandler(http.MethodPost, "/register", AuthRegisterHandler(authSvc)),
		ginutils.WithHandler(http.MethodPost, "/login", AuthLoginHandler(authSvc)),
	)
}

// AuthSendEmailCodeHandler 处理发送邮箱验证码
// @Summary 发送邮箱验证码
// @Description 发送邮箱验证码
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body request.SendEmailCodeReq true "发送邮箱验证码请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /auth/send-email-code [post]
func AuthSendEmailCodeHandler(authSvc ports.AuthService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *request.SendEmailCodeReq) {
		err := authSvc.SendEmailCode(c.Request.Context(), req)
		if handleRouterError(c, err, "auth send email code handler error", errcode.ErrCodeSendEmailCodeFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// AuthRegisterHandler 处理注册
// @Summary 注册
// @Description 注册
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body request.RegisterReq true "注册请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /auth/register [post]
func AuthRegisterHandler(authSvc ports.AuthService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *request.RegisterReq) {
		err := authSvc.Register(c.Request.Context(), req)
		if handleRouterError(c, err, "auth register handler error", errcode.ErrCodeUserRegisterFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// AuthLoginHandler 处理登录
// @Summary 登录
// @Description 登录
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body request.LoginReq true "登录请求体"
// @Success 200 {object} ginutils.Ret[response.TokenResponse]
// @Router /auth/login [post]
func AuthLoginHandler(authSvc ports.AuthService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *request.LoginReq) {
		token, err := authSvc.Login(c.Request.Context(), req)
		if handleRouterError(c, err, "auth login handler error", errcode.ErrCodeUserLoginFailed) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(response.TokenResponse{Token: token}))
	})
}
