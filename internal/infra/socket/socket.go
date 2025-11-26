package socket

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/miebyte/goutils/websocketutils"
)

const (
	BilliardSocketNamespace = "/billiard"
)

// SocketManager 管理 websocket 连接
type SocketManager struct {
	mu                sync.RWMutex
	hook              SessionHook
	socket            websocketutils.ServerAPI
	billiardNamespace websocketutils.NamespaceAPI
	// userConns 管理一个用户多端登录的连接
	userConns  map[uint]map[string]ISession // userID -> sessionID -> ISession
	sessionMap map[string]ISession          // sessionID -> ISession
}

func NewSocketManager(hook SessionHook) *SocketManager {
	socket := websocketutils.NewServer(
		websocketutils.WithNamespacePrefix("/ws"),
		websocketutils.WithHeartbeat(time.Second*10, time.Minute*20),
		websocketutils.WithAllowRequestFunc(hook.OnAllowRequest),
	)

	sm := &SocketManager{
		hook:              hook,
		socket:            socket,
		billiardNamespace: socket.Of(BilliardSocketNamespace),
		userConns:         make(map[uint]map[string]ISession),
		sessionMap:        make(map[string]ISession),
	}
	sm.setupSocket()
	return sm
}
func (sm *SocketManager) serveHttp() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sm.socket.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func (sm *SocketManager) Handler() ginutils.Option {
	return ginutils.WithGroupHandlers(
		ginutils.WithPrefix("/ws"),
		ginutils.WithHandler(http.MethodGet, "", sm.serveHttp()),
		ginutils.WithHandler(http.MethodGet, "/:path", sm.serveHttp()),
	)
}

func (sm *SocketManager) setupSocket() {
	// 用户 websocket 连接成功
	sm.billiardNamespace.On(websocketutils.EventConnection, sm.hook.OnConnect)

	// 用户 websocket 连接断开
	sm.billiardNamespace.On(websocketutils.EventDisconnect, sm.hook.OnDisconnect)
}

func (sm *SocketManager) RegisterSession(userID uint, sessionID string, conn websocketutils.Conn) {
	session := NewSession(sessionID, userID, conn)

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.userConns[userID]; !ok {
		sm.userConns[userID] = make(map[string]ISession)
	}
	sm.userConns[userID][sessionID] = session
	sm.sessionMap[sessionID] = session
}

func (sm *SocketManager) UnregisterSession(conn websocketutils.Conn) {
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

func (sm *SocketManager) GetSessionsByUserID(userID uint) []ISession {
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

func (sm *SocketManager) GetUserSession(userID uint, sessionID string) ISession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if connSet, ok := sm.userConns[userID]; ok {
		return connSet[sessionID]
	}
	return nil
}

func (sm *SocketManager) BroadcastToRoom(ctx context.Context, roomID string, event string, data any) error {
	return sm.billiardNamespace.To(roomID).Emit(event, data)
}
