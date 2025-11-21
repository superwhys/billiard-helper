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
	"github.com/superwhys/billiard-helper/api/middlewares"
	"github.com/superwhys/billiard-helper/models/constant"
	"github.com/superwhys/billiard-helper/models/errcode"
	"github.com/superwhys/billiard-helper/models/response"
	"github.com/superwhys/billiard-helper/pkg/longnet"
	"github.com/superwhys/billiard-helper/service"
)

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

func SocketGroupRouter(sessionManager longnet.ISessionManager, services *service.Service) ginutils.Option {
	socket := websocketutils.NewServer(
		websocketutils.WithPrefix("/ws"),
		websocketutils.WithHeartbeat(time.Second*10, time.Second*20),
		websocketutils.WithHandshake(middlewares.TokenVerifyFromSocket(services.AuthService)),
	)
	setupBilliardSocket(socket, services, sessionManager)

	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/ws"),
		ginutils.WithHandler(http.MethodGet, "", SocketHandler(socket)),
		ginutils.WithHandler(http.MethodGet, "/:path", SocketHandler(socket)),
	)
}

func SocketHandler(socket *websocketutils.Server) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		socket.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func setupBilliardSocket(socket *websocketutils.Server, services *service.Service, sessionManager longnet.ISessionManager) {
	// 创建房间命名空间
	billiardNamespace := socket.Of(constant.BilliardNamespace)
	// 监听该命名空间的 Connection 事件
	billiardNamespace.On(websocketutils.EventConnection, func(s websocketutils.Socket) {
		// 用户连接成功后，将用户连接信息注册到 session manager
		claims, err := middlewares.TokenClaimsFromContext(s.Context())
		if err != nil {
			logging.Errorc(s.Context(), "get token claims from context failed: %v", err)
			return
		}
		sessionManager.RegisterSession(claims.User.ID, s)
		_ = s.Emit(constant.EventClientConnectSuccess, response.ResponseSuccess())
	})
}
