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

func RoomGroupRouter(matchApp *services.MatchApp) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/room"),
		ginutils.WithHandler(http.MethodPost, "/create", RoomCreateHandler(matchApp)),
		ginutils.WithHandler(http.MethodPost, "/join", RoomJoinHandler(matchApp)),
		ginutils.WithHandler(http.MethodPost, "/start", RoomStartHandler(matchApp)),
		ginutils.WithHandler(http.MethodPost, "/end", RoomEndHandler(matchApp)),
		ginutils.WithHandler(http.MethodPost, "/leave", RoomLeaveHandler(matchApp)),
		ginutils.WithHandler(http.MethodPost, "/kick", RoomKickHandler(matchApp)),
	)
}

// RoomCreateHandler 创建房间
// @Summary 创建房间
// @Description 创建房间
// @Tags Room
// @Accept json
// @Produce json
// @Param request body dto.CreateRoomRequest true "创建房间请求体"
// @Success 200 {object} ginutils.Ret[dto.Room]
// @Router /room/create [post]
func RoomCreateHandler(matchApp *services.MatchApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.CreateRoomRequest) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrCodeNoToken) {
			return
		}
		req.UserID = claims.UserID

		ctx := logging.With(c.Request.Context(), "UserID", claims.UserID)
		room, err := matchApp.CreateRoom(ctx, req)
		if handleRouterError(c, err, "create room failed", errcode.ErrCodeCreateRoomFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseWithData(room))
	})
}

// RoomJoinHandler 加入房间
// @Summary 加入房间
// @Description 加入房间
// @Tags Room
// @Accept json
// @Produce json
// @Param request body dto.JoinRoomRequest true "加入房间请求体"
// @Success 200 {object} ginutils.Ret[dto.Room]
// @Router /room/join [post]
func RoomJoinHandler(matchApp *services.MatchApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.JoinRoomRequest) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrCodeNoToken) {
			return
		}
		req.UserID = claims.UserID

		ctx := logging.With(c.Request.Context(), "UserID", claims.UserID)
		room, err := matchApp.JoinRoom(ctx, req)
		if handleRouterError(c, err, "join room failed", errcode.ErrCodeJoinRoomFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseWithData(room))
	})
}

// RoomStartHandler 开始比赛
// @Summary 开始比赛
// @Description 开始比赛
// @Tags Room
// @Accept json
// @Produce json
// @Param request body dto.RoomActionRequest true "开始比赛请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /room/start [post]
func RoomStartHandler(matchApp *services.MatchApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.RoomActionRequest) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrCodeNoToken) {
			return
		}
		req.UserID = claims.UserID

		ctx := logging.With(c.Request.Context(), "UserID", claims.UserID)
		err = matchApp.StartRoom(ctx, req)
		if handleRouterError(c, err, "start room failed", errcode.ErrCodeStartRoomFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// RoomEndHandler 结束比赛
// @Summary 结束比赛
// @Description 结束比赛
// @Tags Room
// @Accept json
// @Produce json
// @Param request body dto.RoomActionRequest true "结束比赛请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /room/end [post]
func RoomEndHandler(matchApp *services.MatchApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.RoomActionRequest) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrCodeNoToken) {
			return
		}
		req.UserID = claims.UserID

		ctx := logging.With(c.Request.Context(), "UserID", claims.UserID)
		err = matchApp.EndRoom(ctx, req)
		if handleRouterError(c, err, "end room failed", errcode.ErrCodeEndRoomFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// RoomLeaveHandler 离开房间
// @Summary 离开房间
// @Description 离开房间
// @Tags Room
// @Accept json
// @Produce json
// @Param request body dto.RoomActionRequest true "离开房间请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /room/leave [post]
func RoomLeaveHandler(matchApp *services.MatchApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.RoomActionRequest) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrCodeNoToken) {
			return
		}
		req.UserID = claims.UserID

		ctx := logging.With(c.Request.Context(), "UserID", claims.UserID)
		err = matchApp.LeaveRoom(ctx, req)
		if handleRouterError(c, err, "leave room failed", errcode.ErrCodeLeaveRoomFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}

// RoomKickHandler 踢出玩家
// @Summary 踢出玩家
// @Description 踢出玩家
// @Tags Room
// @Accept json
// @Produce json
// @Param request body dto.KickPlayerRequest true "踢出玩家请求体"
// @Success 200 {object} ginutils.Ret[any]
// @Router /room/kick [post]
func RoomKickHandler(matchApp *services.MatchApp) gin.HandlerFunc {
	return ginutils.RequestHandler(func(c *gin.Context, req *dto.KickPlayerRequest) {
		claims, err := jwt.TokenClaimsFromContext(c.Request.Context())
		if handleRouterError(c, err, "get token claims failed", errcode.ErrCodeNoToken) {
			return
		}
		req.UserID = claims.UserID

		ctx := logging.With(c.Request.Context(), "UserID", claims.UserID)
		err = matchApp.KickPlayer(ctx, req)
		if handleRouterError(c, err, "kick player failed", errcode.ErrCodeKickPlayerFailed) {
			return
		}

		c.JSON(http.StatusOK, response.ResponseSuccess())
	})
}
