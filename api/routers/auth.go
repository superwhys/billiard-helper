package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/billiard-helper/api/response"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/app/services"
	"github.com/superwhys/billiard-helper/internal/errcode"
)

func AuthGroupRouter(userApp *services.UserApp) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/auth"),
		ginutils.WithHandler(http.MethodPost, "/send-email-code", AuthSendEmailCodeHandler(userApp)),
		ginutils.WithHandler(http.MethodPost, "/register", AuthRegisterHandler(userApp)),
		ginutils.WithHandler(http.MethodPost, "/login", AuthLoginHandler(userApp)),
	)
}

// AuthSendEmailCodeHandler 处理发送邮箱验证码
// @Summary 发送邮箱验证码
// @Description 发送邮箱验证码
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.SendRegisterCodeReq true "发送邮箱验证码请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /auth/send-email-code [post]
func AuthSendEmailCodeHandler(userApp *services.UserApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.SendRegisterCodeReq) {
		err := userApp.SendRegisterCode(c.Request.Context(), req)
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
// @Param request body dto.RegisterReq true "注册请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /auth/register [post]
func AuthRegisterHandler(userApp *services.UserApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.RegisterReq) {
		err := userApp.Register(c.Request.Context(), req)
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
// @Param request body dto.LoginReq true "登录请求体"
// @Success 200 {object} ginutils.Ret[dto.TokenResponse]
// @Router /auth/login [post]
func AuthLoginHandler(userApp *services.UserApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.LoginReq) {
		token, user, err := userApp.Login(c.Request.Context(), req)
		if handleRouterError(c, err, "auth login handler error", errcode.ErrCodeUserLoginFailed) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(dto.TokenResponse{
			Token: token,
			User:  user,
		}))
	})
}
