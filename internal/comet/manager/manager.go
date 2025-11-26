package manager

import (
	"context"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/websocketutils"
	"github.com/superwhys/billiard-helper/api/middlewares"
	"github.com/superwhys/billiard-helper/internal/models/constant"
	"github.com/superwhys/billiard-helper/internal/models/request"
	"github.com/superwhys/billiard-helper/internal/models/response"
	"github.com/superwhys/billiard-helper/internal/models/types"
	"github.com/superwhys/billiard-helper/internal/service"
)

// SessionManager 管理 websocket 连接
type SessionManager struct {
	mu                sync.RWMutex
	srv               *service.Service
	socket            websocketutils.ServerAPI
	billiardNamespace websocketutils.NamespaceAPI
	// userConns 管理一个用户多端登录的连接
	userConns  map[uint]map[string]ISession // userID -> sessionID -> ISession
	sessionMap map[string]ISession          // sessionID -> ISession
}

func NewSessionManager(srv *service.Service) *SessionManager {
	socket := websocketutils.NewServer(
		websocketutils.WithNamespacePrefix("/ws"),
		websocketutils.WithHeartbeat(time.Second*10, time.Minute*20),
		websocketutils.WithAllowRequestFunc(middlewares.TokenVerifyFromSocket(srv.AuthService)),
	)

	sm := &SessionManager{
		srv:               srv,
		socket:            socket,
		billiardNamespace: socket.Of(constant.BilliardSocketNamespace),
		userConns:         make(map[uint]map[string]ISession),
		sessionMap:        make(map[string]ISession),
	}
	sm.setupSocket()
	return sm
}

func (sm *SessionManager) setupSocket() {
	// 用户 websocket 连接成功
	sm.billiardNamespace.On(websocketutils.EventConnection, func(ctx *websocketutils.Context) {
		claims, err := middlewares.TokenClaimsFromContext(ctx.Context())
		if err != nil {
			logging.Errorc(ctx.Context(), "get token claims from context failed: %v", err)
			return
		}

		sm.RegisterSession(claims.User.ID, claims.SessionID, ctx.Conn())
		_ = ctx.Conn().Emit(constant.EventClientConnectSuccess, response.ResponseWithData(claims))
	})

	// 用户 websocket 连接断开
	sm.billiardNamespace.On(websocketutils.EventDisconnect, func(ctx *websocketutils.Context) {
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
			sm.srv.RoomService.LeaveRoom(ctx.Context(), &request.LeaveRoomRequest{
				RoomID:      types.ParseRoomID(room),
				UserID:      claims.User.ID,
				PlayerCode:  room,
				ReallyLeave: false,
			})
		}
	})
}

func (sm *SessionManager) ServeHttp() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sm.socket.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func (sm *SessionManager) RegisterSession(userID uint, sessionID string, conn websocketutils.Conn) {
	session := NewSession(sessionID, userID, conn)

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.userConns[userID]; !ok {
		sm.userConns[userID] = make(map[string]ISession)
	}
	sm.userConns[userID][sessionID] = session
	sm.sessionMap[sessionID] = session
}

func (sm *SessionManager) UnregisterSession(conn websocketutils.Conn) {
	connID := conn.ID()

	sm.mu.Lock()
	defer sm.mu.Unlock()

	var (
		session   ISession
		sessionID string
	)

	for sid, s := range sm.sessionMap {
		if s.ConnID() == connID {
			session = s
			sessionID = sid
			break
		}
	}

	if session == nil {
		return
	}

	delete(sm.sessionMap, sessionID)

	if connSet, ok := sm.userConns[session.UserID()]; ok {
		delete(connSet, sessionID)
		if len(connSet) == 0 {
			delete(sm.userConns, session.UserID())
		}
	}
}

func (sm *SessionManager) GetSessionsByUserID(userID uint) []ISession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if connSet, ok := sm.userConns[userID]; ok {
		sessions := make([]ISession, 0, len(connSet))
		for _, session := range connSet {
			sessions = append(sessions, session)
		}
		return sessions
	}
	return nil
}

func (sm *SessionManager) GetUserSession(userID uint, sessionID string) ISession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if connSet, ok := sm.userConns[userID]; ok {
		return connSet[sessionID]
	}
	return nil
}

func (sm *SessionManager) BroadcastToRoom(ctx context.Context, roomID string, event string, data any) error {
	return sm.billiardNamespace.To(roomID).Emit(event, data)
}
