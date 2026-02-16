package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/billiard-helper/api/response"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/app/services"
	"github.com/superwhys/billiard-helper/internal/errcode"
	"github.com/superwhys/billiard-helper/internal/pkg/jwt"
)

func FeedbackRouter(feedbackApp *services.FeedbackApp) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/feedback"),
		ginutils.WithHandler(http.MethodPost, "/report", FeedbackReportHandler(feedbackApp)),
	)
}

// FeedbackReportHandler 反馈
// @Summary 反馈
// @Description 反馈
// @Tags Feedback
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.FeedbackReportReq true "反馈请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /feedback/report [post]
func FeedbackReportHandler(feedbackApp *services.FeedbackApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.FeedbackReportReq) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrUnauthorized) {
			return
		}
		req.UserID = claims.UserID

		ctx := logging.With(c.Request.Context(), "UserID", claims.UserID)
		err = feedbackApp.ReportFeedback(ctx, req)
		if handleRouterError(c, err, "report feedback failed", errcode.ErrCodeReportFeedbackFailed) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}
