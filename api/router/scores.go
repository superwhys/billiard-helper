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

func ScoresGroupRouter(scoresSvc ports.ScoresService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/scores"),
		ginutils.WithHandler(http.MethodGet, "/list", ScoresListHandler(scoresSvc)),
	)
}

// ScoresListHandler 处理获取房间分数列表
// @Summary 获取房间分数列表
// @Description 获取房间分数列表
// @Tags Scores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.GetRoomScoresRequest true "获取房间分数列表请求体"
// @Success 200 {object} ginutils.Ret[[]response.Scores]
// @Router /scores/list [get]
func ScoresListHandler(scoresSvc ports.ScoresService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *request.GetRoomScoresRequest) {
		scores, err := scoresSvc.GetRoomScores(c.Request.Context(), req)
		if handleRouterError(c, err, "scores list handler error", errcode.ErrCodeGetRoomScoresFailed) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseWithData(scores))
	})
}
