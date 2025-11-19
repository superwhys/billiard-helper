package router

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/websocketutils"
	"github.com/superwhys/billiard-helper/models/constant"
	"github.com/superwhys/billiard-helper/models/errcode"
	"github.com/superwhys/billiard-helper/models/request"
	"github.com/superwhys/billiard-helper/models/response"
	"github.com/superwhys/billiard-helper/models/types"
	"github.com/superwhys/billiard-helper/service"
)

func SocketGroupRouter(services *service.Service) ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/ws"),
		ginutils.WithHandler(http.MethodGet, "/", SocketHandler(services)),
	)
}

func SocketHandler(services *service.Service) gin.HandlerFunc {
	socket := websocketutils.NewServer(
		websocketutils.WithHeartbeat(time.Second*10, time.Second*20),
	)
	setupBilliardSocket(services, socket)
	return func(ctx *gin.Context) {
		socket.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func setupBilliardSocket(services *service.Service, socket *websocketutils.Server) {
	billiardNamespace := socket.Of(constant.BilliardNamespace)
	billiardNamespace.On(websocketutils.EventConnection, func(s websocketutils.Socket) {
		s.On(constant.EventClientJoinRoom, socketEventHandler(joinRoomEventHandler(services, billiardNamespace)))
		s.On(constant.EventClientLeaveRoom, socketEventHandler(leaveRoomEventHandler(services, billiardNamespace)))

		s.On(constant.EventClientScoreAdd, func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On(constant.EventClientScoreMinus, func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On(constant.EventClientScoreReset, func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On(constant.EventClientKickPlayer, func(s websocketutils.Socket, rm json.RawMessage) {})
	})
}

func joinRoomEventHandler(services *service.Service, billiardNamespace *websocketutils.Namespace) eventHandlerFunc[*request.JoinRoomRequest] {
	return func(ctx context.Context, s websocketutils.Socket, req *eventData[*request.JoinRoomRequest]) {
		room, err := services.RoomService.JoinRoom(ctx, req.Payload)
		if err != nil {
			logging.Errorc(ctx, "join room failed: %v", err)
			s.Emit(constant.EventCallbackJoinRoomFailed, errorResponseWithCode(err, errcode.ErrCodeJoinRoomFailed))
			return
		}

		logging.Infoc(ctx, "join room success: %v", room)
		callbackData := generateEventCallbackData(req.RequestID, time.Now().Unix(), room)
		s.Emit(constant.EventCallbackJoinRoomSuccess, response.ResponseWithData(callbackData))
		s.Join(room.SocketRoomID())
		billiardNamespace.To(room.SocketRoomID()).Emit(constant.EventBroadcastRoomJoin, response.ResponseWithData(room))
	}
}

func leaveRoomEventHandler(services *service.Service, billiardNamespace *websocketutils.Namespace) eventHandlerFunc[*request.LeaveRoomRequest] {
	return func(ctx context.Context, s websocketutils.Socket, req *eventData[*request.LeaveRoomRequest]) {
		err := services.RoomService.LeaveRoom(ctx, req.Payload)
		if err != nil {
			logging.Errorc(ctx, "leave room failed: %v", err)
			s.Emit(constant.EventCallbackLeaveRoomFailed, errorResponseWithCode(err, errcode.ErrCodeLeaveRoomFailed))
		}

		logging.Infoc(ctx, "leave room success: %v", req.Payload.PlayerCode)
		callbackData := generateEventCallbackData(req.RequestID, time.Now().Unix(), req.Payload.PlayerCode)

		s.Emit(constant.EventCallbackLeaveRoomSuccess, response.ResponseWithData(callbackData))
		billiardNamespace.
			To(types.SocketRoomID(req.Payload.RoomID)).
			Emit(constant.EventBroadcastRoomLeave, response.ResponseWithData(req.Payload.PlayerCode))
		s.Leave(types.SocketRoomID(req.Payload.RoomID))
	}
}

type eventData[R any] struct {
	RequestID string `json:"request_id"`
	Payload   R      `json:"payload"`
	Timestamp int64  `json:"timestamp"`
}

type eventCallbackData struct {
	RequestID string `json:"request_id"`
	Timestamp int64  `json:"timestamp"`
	Data      any    `json:"data"`
}

func generateEventCallbackData(requestID string, timestamp int64, data any) eventCallbackData {
	return eventCallbackData{
		RequestID: requestID,
		Timestamp: timestamp,
		Data:      data,
	}
}

type eventHandlerFunc[R any] func(ctx context.Context, s websocketutils.Socket, req *eventData[R])

func socketEventHandler[R any](fn eventHandlerFunc[R]) websocketutils.MessageHandler {
	return func(s websocketutils.Socket, rm json.RawMessage) {
		var req eventData[R]
		err := json.Unmarshal(rm, &req)
		if err != nil {
			logging.Errorc(s.Context(), "unmarshal request failed: %v", err)
			s.Emit(constant.EventCallbackFailed, response.ErrorResponseWithCode(errcode.ErrCodeInvalidRequest))
			return
		}

		ctx := logging.With(s.Context(), "RequestID", req.RequestID, "Timestamp", req.Timestamp)
		fn(ctx, s, &req)
	}
}
