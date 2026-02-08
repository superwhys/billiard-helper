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

func MatchGroupRouter(matchApp *services.MatchApp) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/match"),
		ginutils.WithHandler(http.MethodGet, "/list", MatchListHandler(matchApp)),
		ginutils.WithHandler(http.MethodGet, "/detail", MatchDetailHandler(matchApp)),
		ginutils.WithHandler(http.MethodPost, "/create", MatchCreateHandler(matchApp)),
		ginutils.WithHandler(http.MethodPost, "/join", MatchJoinHandler(matchApp)),
		ginutils.WithHandler(http.MethodPost, "/start", MatchStartHandler(matchApp)),
		ginutils.WithHandler(http.MethodPost, "/end", MatchEndHandler(matchApp)),
		ginutils.WithHandler(http.MethodPost, "/leave", MatchLeaveHandler(matchApp)),
		ginutils.WithHandler(http.MethodPost, "/kick", MatchKickHandler(matchApp)),
	)
}

// MatchListHandler 获取比赛列表
// @Summary 获取比赛列表
// @Description 获取比赛列表
// @Tags Match
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.MatchListRequest true "获取比赛列表请求体"
// @Success 200 {object} ginutils.Ret[[]dto.Match]
// @Router /match/list [get]
func MatchListHandler(matchApp *services.MatchApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.MatchListRequest) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrUnauthorized) {
			return
		}

		matches, err := matchApp.ListMatches(c.Request.Context(), claims.UserID, req)
		if handleRouterError(c, err, "list matches failed", errcode.ErrCodeListMatchesFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseWithData(matches))
	})
}

// MatchDetailHandler 获取比赛详情
// @Summary 获取比赛详情
// @Description 获取比赛详情
// @Tags Match
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.MatchDetailRequest true "获取比赛详情请求体"
// @Success 200 {object} ginutils.Ret[dto.Match]
// @Router /match/detail [get]
func MatchDetailHandler(matchApp *services.MatchApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.MatchDetailRequest) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrUnauthorized) {
			return
		}

		match, err := matchApp.GetMatchDetail(c.Request.Context(), claims.UserID, req.MatchID)
		if handleRouterError(c, err, "get match detail failed", errcode.ErrCodeMatchDetailFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseWithData(match))
	})
}

// MatchCreateHandler 创建房间
// @Summary 创建房间
// @Description 创建房间
// @Tags Match
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateMatchRequest true "创建比赛请求体"
// @Success 200 {object} ginutils.Ret[dto.Match]
// @Router /match/create [post]
func MatchCreateHandler(matchApp *services.MatchApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.CreateMatchRequest) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrUnauthorized) {
			return
		}
		req.UserID = claims.UserID

		ctx := logging.With(c.Request.Context(), "UserID", claims.UserID)
		Match, err := matchApp.CreateMatch(ctx, req)
		if handleRouterError(c, err, "create Match failed", errcode.ErrCodeCreateMatchFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseWithData(Match))
	})
}

// MatchJoinHandler 加入房间
// @Summary 加入房间
// @Description 加入房间
// @Tags Match
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.JoinMatchRequest true "加入房间请求体"
// @Success 200 {object} ginutils.Ret[dto.Match]
// @Router /match/join [post]
func MatchJoinHandler(matchApp *services.MatchApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.JoinMatchRequest) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrUnauthorized) {
			return
		}
		req.UserID = claims.UserID

		ctx := logging.With(c.Request.Context(), "UserID", claims.UserID)
		Match, err := matchApp.JoinMatch(ctx, req)
		if handleRouterError(c, err, "join Match failed", errcode.ErrCodeJoinMatchFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseWithData(Match))
	})
}

// MatchStartHandler 开始比赛
// @Summary 开始比赛
// @Description 开始比赛
// @Tags Match
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.MatchActionRequest true "开始比赛请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /match/start [post]
func MatchStartHandler(matchApp *services.MatchApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.MatchActionRequest) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrUnauthorized) {
			return
		}
		req.UserID = claims.UserID

		ctx := logging.With(c.Request.Context(), "UserID", claims.UserID)
		err = matchApp.StartMatch(ctx, req)
		if handleRouterError(c, err, "start Match failed", errcode.ErrCodeStartMatchFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// MatchEndHandler 结束比赛
// @Summary 结束比赛
// @Description 结束比赛
// @Tags Match
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.MatchActionRequest true "结束比赛请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /match/end [post]
func MatchEndHandler(matchApp *services.MatchApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.MatchActionRequest) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrUnauthorized) {
			return
		}
		req.UserID = claims.UserID

		ctx := logging.With(c.Request.Context(), "UserID", claims.UserID)
		err = matchApp.EndMatch(ctx, req)
		if handleRouterError(c, err, "end Match failed", errcode.ErrCodeEndMatchFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// MatchLeaveHandler 离开房间
// @Summary 离开房间
// @Description 离开房间
// @Tags Match
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.MatchActionRequest true "离开房间请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /match/leave [post]
func MatchLeaveHandler(matchApp *services.MatchApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.MatchActionRequest) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrUnauthorized) {
			return
		}
		req.UserID = claims.UserID

		ctx := logging.With(c.Request.Context(), "UserID", claims.UserID)
		err = matchApp.LeaveMatch(ctx, req)
		if handleRouterError(c, err, "leave Match failed", errcode.ErrCodeLeaveMatchFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// MatchKickHandler 踢出玩家
// @Summary 踢出玩家
// @Description 踢出玩家
// @Tags Match
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.KickPlayerRequest true "踢出玩家请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /match/kick [post]
func MatchKickHandler(matchApp *services.MatchApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.KickPlayerRequest) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrUnauthorized) {
			return
		}
		req.UserID = claims.UserID

		ctx := logging.With(c.Request.Context(), "UserID", claims.UserID)
		err = matchApp.KickMatchPlayer(ctx, req)
		if handleRouterError(c, err, "kick player failed", errcode.ErrCodeKickPlayerFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}
