package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/billiard-helper/models/errcode"
	"github.com/superwhys/billiard-helper/models/response"
)

// handleRouterError 处理业务逻辑错误响应。
func handleRouterError(ctx *gin.Context, err error, logMsg string, fallback errcode.ErrCode) bool {
	if err == nil {
		return false
	}
	if ec, ok := errcode.AsErrcode(err); ok {
		ctx.JSON(http.StatusOK, response.ErrorResponseWithCode(ec))
		return true
	}
	logging.Errorc(ctx, "%s: %v", logMsg, err)
	ctx.JSON(http.StatusOK, response.ErrorResponseWithCode(fallback))
	return true
}
