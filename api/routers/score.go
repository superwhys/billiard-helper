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

func ScoreGroupRouter(scoreApp *services.ScoreApp) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/score"),
		ginutils.WithHandler(http.MethodPost, "/sync", ScoreSyncHandler(scoreApp)),
		ginutils.WithHandler(http.MethodPost, "/undo", ScoreUndoHandler(scoreApp)),
		ginutils.WithHandler(http.MethodGet, "/list", ScoreListHandler(scoreApp)),
	)
}

// ScoreSyncHandler 同步分数
// @Summary 同步分数
// @Description 同步分数
// @Tags Score
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.MatchScoreSyncEvent true "同步分数请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /score/sync [post]
func ScoreSyncHandler(scoreApp *services.ScoreApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.MatchScoreSyncEvent) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrUnauthorized) {
			return
		}
		req.UserID = claims.UserID

		ctx := logging.With(c.Request.Context(), "UserID", claims.UserID)
		newScores, err := scoreApp.SyncScore(ctx, req)
		if handleRouterError(c, err, "sync score failed", errcode.ErrCodeSyncScoreFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseWithData(newScores))
	})
}

// ScoreUndoHandler 撤回分数
// @Summary 撤回分数
// @Description 撤回分数
// @Tags Score
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.MatchScoreUndoReq true "撤回分数请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /score/undo [post]
func ScoreUndoHandler(scoreApp *services.ScoreApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.MatchScoreUndoReq) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrUnauthorized) {
			return
		}
		req.UserID = claims.UserID

		ctx := logging.With(c.Request.Context(), "UserID", claims.UserID)
		newScores, err := scoreApp.UndoScore(ctx, req)
		if handleRouterError(c, err, "undo score failed", errcode.ErrCodeUndoScoreFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseWithData(newScores))
	})
}

// ScoreListHandler 获取分数历史记录
// @Summary 获取分数历史记录
// @Description 获取分数历史记录
// @Tags Score
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.MatchScoreListReq true "获取分数历史记录请求体"
// @Success 200 {object} ginutils.Ret[[]dto.MatchScoreSyncEvent]
// @Router /score/list [get]
func ScoreListHandler(scoreApp *services.ScoreApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.MatchScoreListReq) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrUnauthorized) {
			return
		}
		req.UserID = claims.UserID

		ctx := logging.With(c.Request.Context(), "UserID", claims.UserID)
		scores, err := scoreApp.ListScores(ctx, req)
		if handleRouterError(c, err, "list scores failed", errcode.ErrCodeListScoresFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseWithData(scores))
	})
}
