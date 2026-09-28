package socket

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/ginutils"
	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/websocketutils"
	"github.com/superwhys/billiard-helper/api/response"
	"github.com/superwhys/billiard-helper/internal/constant"
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
	userConns         map[uint]map[string]ISession // userID -> connID -> ISession
	connIDMap         map[string]uint              // connID -> userID
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
		connIDMap:         make(map[string]uint),
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
		ginutils.WithHandler(http.MethodGet, BilliardSocketNamespace, sm.serveHttp()),
	)
}

func (sm *SocketManager) setupSocket() {
	// 用户 websocket 连接成功
	sm.billiardNamespace.On(websocketutils.EventConnection, func(ctx *websocketutils.Context) {
		userID, err := sm.hook.OnConnect(ctx.Context())
		if err != nil {
			logging.Errorc(ctx.Context(), "on connect failed: %v", err)
			_ = ctx.Conn().Close()
			return
		}

		sm.RegisterSession(userID, ctx.Conn())
		_ = ctx.Conn().Emit(constant.EventClientConnectSuccess, response.ResponseWithData(userID))

	})

	// 用户 websocket 连接断开
	sm.billiardNamespace.On(websocketutils.EventDisconnect, func(ctx *websocketutils.Context) {
		sm.hook.OnDisconnect(ctx.Context(), ctx.Conn())
		sm.UnregisterSession(ctx.Conn())
	})
}

func (sm *SocketManager) RegisterSession(userID uint, conn websocketutils.Conn) {
	session := NewSession(userID, conn)

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.userConns[userID] == nil {
		sm.userConns[userID] = make(map[string]ISession)
	}
	sm.userConns[userID][conn.ID()] = session
	sm.connIDMap[conn.ID()] = userID
}

func (sm *SocketManager) UnregisterSession(conn websocketutils.Conn) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	userID, ok := sm.connIDMap[conn.ID()]
	if !ok {
		return
	}
	delete(sm.userConns[userID], conn.ID())
	if len(sm.userConns[userID]) == 0 {
		delete(sm.userConns, userID)
	}
	delete(sm.connIDMap, conn.ID())
}

func (sm *SocketManager) userSessions(userID uint) []ISession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]ISession, 0, len(sm.userConns[userID]))
	for _, session := range sm.userConns[userID] {
		sessions = append(sessions, session)
	}
	return sessions
}

func (sm *SocketManager) GetUserSessions(ctx context.Context, userID uint) []ISession {
	sessions := sm.userSessions(userID)
	if len(sessions) == 0 {
		return nil
	}

	blocked, err := sm.hook.IsUserBlocked(ctx, userID)
	if err != nil || blocked {
		if err != nil {
			logging.Errorc(ctx, "check user(%d) blocked status failed: %v", userID, err)
		}
		sm.CloseUserConnections(userID)
		return nil
	}
	return sessions
}

func (sm *SocketManager) CloseUserConnections(userID uint) {
	for _, session := range sm.userSessions(userID) {
		if err := session.Close(); err != nil {
			logging.Errorf("close user(%d) websocket connection failed: %v", userID, err)
		}
	}
}

func (sm *SocketManager) BroadcastToRoom(ctx context.Context, roomID string, event string, data any) error {
	var firstErr error
	for _, conn := range sm.billiardNamespace.Room(roomID).Members() {
		sm.mu.RLock()
		userID, ok := sm.connIDMap[conn.ID()]
		sm.mu.RUnlock()
		if !ok {
			_ = conn.Close()
			continue
		}

		blocked, err := sm.hook.IsUserBlocked(ctx, userID)
		if err != nil || blocked {
			if err != nil {
				logging.Errorc(ctx, "check user(%d) blocked status failed: %v", userID, err)
				if firstErr == nil {
					firstErr = err
				}
			}
			_ = conn.Close()
			continue
		}
		if err := conn.Emit(event, data); err != nil {
			logging.Errorc(ctx, "room broadcast failed room=%s conn=%s event=%s err=%v", roomID, conn.ID(), event, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}
