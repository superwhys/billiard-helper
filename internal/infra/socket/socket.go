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
	userConns         map[uint]ISession // userID -> ISession
	connIDMap         map[string]uint   // connID -> userID
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
		userConns:         make(map[uint]ISession),
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
		ginutils.WithHandler(http.MethodGet, "", sm.serveHttp()),
		ginutils.WithHandler(http.MethodGet, "/:path", sm.serveHttp()),
	)
}

func (sm *SocketManager) setupSocket() {
	// 用户 websocket 连接成功
	sm.billiardNamespace.On(websocketutils.EventConnection, func(ctx *websocketutils.Context) {
		userID, err := sm.hook.OnConnect(ctx.Context())
		if err != nil {
			logging.Errorc(ctx.Context(), "on connect failed: %v", err)
			return
		}

		sm.RegisterSession(userID, ctx.Conn())
		_ = ctx.Conn().Emit(constant.EventClientConnectSuccess, response.ResponseWithData(userID))

	})

	// 用户 websocket 连接断开
	sm.billiardNamespace.On(websocketutils.EventDisconnect, func(ctx *websocketutils.Context) {
		sm.hook.OnDisconnect(ctx.Context(), ctx.Conn())
	})
}

func (sm *SocketManager) RegisterSession(userID uint, conn websocketutils.Conn) {
	session := NewSession(userID, conn)

	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, ok := sm.userConns[userID]; !ok {
		sm.userConns[userID] = session
		sm.connIDMap[conn.ID()] = userID
	}
}

func (sm *SocketManager) UnregisterSession(conn websocketutils.Conn) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.userConns, sm.connIDMap[conn.ID()])
	delete(sm.connIDMap, conn.ID())
}

func (sm *SocketManager) GetUserSession(userID uint) ISession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.userConns[userID]
}

func (sm *SocketManager) BroadcastToRoom(ctx context.Context, roomID string, event string, data any) error {
	return sm.billiardNamespace.To(roomID).Emit(event, data)
}
