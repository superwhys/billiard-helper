package routers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/billiard-helper/api/response"
	"github.com/superwhys/billiard-helper/internal/errcode"
)

// handleRouterError 处理业务逻辑错误响应。
func handleRouterError(ctx *gin.Context, err error, logMsg string, fallback errcode.Error) bool {
	if err == nil {
		return false
	}
	logging.Errorc(ctx, "%s: %v", logMsg, err)
	status := http.StatusOK
	if errors.Is(err, errcode.ErrForbidden) {
		status = http.StatusForbidden
	}
	ctx.JSON(status, errorResponseWithCode(err, fallback))
	return true
}

func errorResponseWithCode(err error, fallback errcode.Error) *ginutils.Ret[any] {
	var ec errcode.Error
	if errors.As(err, &ec) {
		return response.ErrorResponseWithCode(ec)
	}
	return response.ErrorResponseWithCode(fallback)
}
