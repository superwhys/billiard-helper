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
	"github.com/superwhys/billiard-helper/internal/models/request"
	"github.com/superwhys/billiard-helper/internal/models/response"
	"github.com/superwhys/billiard-helper/internal/models/types"
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
	// 用户 websocket 连接成功
	billiardNamespace.On(websocketutils.EventConnection, func(ctx *websocketutils.Context) {
		claims, err := middlewares.TokenClaimsFromContext(ctx.Context())
		if err != nil {
			logging.Errorc(ctx.Context(), "get token claims from context failed: %v", err)
			return
		}

		s.sessionManager.RegisterSession(claims.User.ID, claims.SessionID, ctx.Conn())
		_ = ctx.Conn().Emit(constant.EventClientConnectSuccess, response.ResponseWithData(claims))
	})

	// 用户 websocket 连接断开
	billiardNamespace.On(websocketutils.EventDisconnect, func(ctx *websocketutils.Context) {
		claims, err := middlewares.TokenClaimsFromContext(ctx.Context())
		if err != nil {
			logging.Errorc(ctx.Context(), "get token claims from context failed: %v", err)
			return
		}

		logging.Infoc(ctx.Context(), "user(%d) disconnect", claims.User.ID)

		// 通知所有加入的房间。
		enterRooms := ctx.Conn().Rooms()
		for _, room := range enterRooms {
			if !types.IsSocketRoomID(room) {
				continue
			}
			logging.Infoc(ctx.Context(), "user(%d) leave room(%s)", claims.User.ID, room)
			s.srv.RoomService.LeaveRoom(ctx.Context(), &request.LeaveRoomRequest{
				RoomID:      types.ParseRoomID(room),
				UserID:      claims.User.ID,
				PlayerCode:  room,
				ReallyLeave: false,
			})
		}
	})
}
