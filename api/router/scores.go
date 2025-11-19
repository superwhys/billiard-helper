package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/billiard-helper/models/errcode"
	"github.com/superwhys/billiard-helper/models/request"
	"github.com/superwhys/billiard-helper/models/response"
	"github.com/superwhys/billiard-helper/ports"
)

func ScoresGroupRouter(scoresSvc ports.ScoresService) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/scores"),
		ginutils.WithHandler(http.MethodPost, "/add", ScoresAddHandler(scoresSvc)),
		ginutils.WithHandler(http.MethodPost, "/minus", ScoresMinusHandler(scoresSvc)),
		ginutils.WithHandler(http.MethodPost, "/reset", ScoresResetHandler(scoresSvc)),
		ginutils.WithHandler(http.MethodGet, "/list", ScoresListHandler(scoresSvc)),
	)
}

// ScoresAddHandler 处理添加分数
// @Summary 添加分数
// @Description 添加分数
// @Tags Scores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.AddScoreRequest true "添加分数请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /scores/add [post]
func ScoresAddHandler(scoresSvc ports.ScoresService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *request.AddScoreRequest) {
		err := scoresSvc.AddScore(c.Request.Context(), req)
		if handleRouterError(c, err, "scores add handler error", errcode.ErrCodeAddScoreFailed) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// ScoresMinusHandler 处理减少分数
// @Summary 减少分数
// @Description 减少分数
// @Tags Scores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.MinusScoreRequest true "减少分数请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /scores/minus [post]
func ScoresMinusHandler(scoresSvc ports.ScoresService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *request.MinusScoreRequest) {
		err := scoresSvc.MinusScore(c.Request.Context(), req)
		if handleRouterError(c, err, "scores minus handler error", errcode.ErrCodeMinusScoreFailed) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// ScoresResetHandler 处理重置分数
// @Summary 重置分数
// @Description 重置分数
// @Tags Scores
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.ResetScoreRequest true "重置分数请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /scores/reset [post]
func ScoresResetHandler(scoresSvc ports.ScoresService) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *request.ResetScoreRequest) {
		err := scoresSvc.ResetScore(c.Request.Context(), req)
		if handleRouterError(c, err, "scores reset handler error", errcode.ErrCodeResetScoreFailed) {
			return
		}
		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
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
