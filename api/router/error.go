package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/billiard-helper/internal/models/errcode"
	"github.com/superwhys/billiard-helper/internal/models/response"
)

// handleRouterError 处理业务逻辑错误响应。
func handleRouterError(ctx *gin.Context, err error, logMsg string, fallback errcode.ErrCode) bool {
	if err == nil {
		return false
	}
	logging.Errorc(ctx, "%s: %v", logMsg, err)
	ctx.JSON(http.StatusOK, errorResponseWithCode(err, fallback))
	return true
}

func errorResponseWithCode(err error, fallback errcode.ErrCode) *ginutils.Ret[any] {
	if ec, ok := errcode.AsErrcode(err); ok {
		return response.ErrorResponseWithCode(ec)
	}
	return response.ErrorResponseWithCode(fallback)
}
