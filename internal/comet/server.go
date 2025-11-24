package comet

import (
	"context"
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

type Server struct {
	srv            *service.Service
	subscriber     *Subscriber
	socket         websocketutils.ServerAPI
	sessionManager longnet.ISessionManager
}

func NewCometServer(queue longnet.EventQueue, sessionManager longnet.ISessionManager, srv *service.Service) *Server {
	subscriber := NewSubscriber(queue, sessionManager, srv)

	socket := websocketutils.NewServer(
		websocketutils.WithNamespacePrefix("/ws"),
		websocketutils.WithHeartbeat(time.Second*10, time.Minute*20),
		websocketutils.WithAllowRequestFunc(middlewares.TokenVerifyFromSocket(srv.AuthService)),
	)

	server := &Server{
		srv:            srv,
		subscriber:     subscriber,
		socket:         socket,
		sessionManager: sessionManager,
	}

	server.setupSocket()

	return server
}

func (s *Server) Subscribe(ctx context.Context) error {
	return s.subscriber.Subscribe(ctx)
}

func (s *Server) handlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		s.socket.ServeHTTP(ctx.Writer, ctx.Request)
	}
}
func (s *Server) Handler() ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/ws"),
		ginutils.WithHandler(http.MethodGet, "", s.handlerFunc()),
		ginutils.WithHandler(http.MethodGet, "/:path", s.handlerFunc()),
	)
}

func (s *Server) setupSocket() {
	billiardNamespace := s.socket.Of(constant.BilliardNamespace)
	billiardNamespace.On(websocketutils.EventConnection, func(ctx *websocketutils.Context) {
		// 用户连接成功后，将用户连接信息注册到 session manager
		claims, err := middlewares.TokenClaimsFromContext(ctx.Context())
		if err != nil {
			logging.Errorc(ctx.Context(), "get token claims from context failed: %v", err)
			return
		}
		s.sessionManager.RegisterSession(claims, ctx.Conn())
		_ = ctx.Conn().Emit(constant.EventClientConnectSuccess, response.ResponseWithData(claims))
	})
}
