package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/websocketutils"
	"github.com/superwhys/billiard-helper/api/middlewares"
	"github.com/superwhys/billiard-helper/internal/models/constant"
	"github.com/superwhys/billiard-helper/internal/models/response"
	"github.com/superwhys/billiard-helper/internal/pkg/longnet"
	"github.com/superwhys/billiard-helper/internal/service"
)

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
