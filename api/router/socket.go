package router

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/websocketutils"
	"github.com/superwhys/billiard-helper/models/constant"
	"github.com/superwhys/billiard-helper/models/errcode"
	"github.com/superwhys/billiard-helper/models/request"
	"github.com/superwhys/billiard-helper/models/response"
	"github.com/superwhys/billiard-helper/service"
)

func SocketGroupRouter(services *service.Service) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/ws"),
		ginutils.WithHandler(http.MethodGet, "/", SocketHandler(services)),
	)
}

func SocketHandler(services *service.Service) gin.HandlerFunc {
	socket := resolveSocketServer(services)
	setupBilliardSocket(services, socket)
	return func(ctx *gin.Context) {
		socket.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func setupBilliardSocket(services *service.Service, socket *websocketutils.Server) {
	billiardNamespace := socket.Of(constant.BilliardNamespace)
	billiardNamespace.On(websocketutils.EventConnection, func(s websocketutils.Socket) {
		s.On(constant.EventClientJoinRoom, func(s websocketutils.Socket, rm json.RawMessage) {
			joinReq := &request.JoinRoomRequest{}
			err := json.Unmarshal(rm, joinReq)
			if err != nil {
				logging.Errorc(s.Context(), "unmarshal join room request failed: %v", err)
				s.Emit(constant.EventCallbackJoinRoomFailed, response.ErrorResponseWithCode(errcode.ErrCodeInvalidRequest))
				return
			}

			room, err := services.RoomService.JoinRoom(s.Context(), joinReq)
			if err != nil {
				logging.Errorc(s.Context(), "join room failed: %v", err)
				s.Emit(constant.EventCallbackJoinRoomFailed, errorResponseWithCode(err, errcode.ErrCodeJoinRoomFailed))
				return
			}

			logging.Infoc(s.Context(), "join room success: %v", room)
			s.Emit(constant.EventCallbackJoinRoomSuccess, response.ResponseWithData(room))
			s.Join(room.SocketRoomID())

			// TODO: 广播玩家加入房间的事件, 应该广播加入的玩家信息
			billiardNamespace.To(room.SocketRoomID()).Emit(constant.EventBroadcastRoomJoin, response.ResponseWithData(room))
		})

		s.On(constant.EventClientLeaveRoom, func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On(constant.EventClientScoreAdd, func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On(constant.EventClientScoreMinus, func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On(constant.EventClientScoreReset, func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On(constant.EventClientKickPlayer, func(s websocketutils.Socket, rm json.RawMessage) {})
	})
}

func resolveSocketServer(services *service.Service) *websocketutils.Server {
	if services != nil {
		if ctx := services.Context(); ctx != nil && ctx.Socket != nil {
			return ctx.Socket
		}
	}
	return websocketutils.NewServer()
}
