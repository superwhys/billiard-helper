package longnet

import (
	"context"
	"sync"

	"github.com/miebyte/goutils/websocketutils"
	"github.com/superwhys/billiard-helper/internal/models/errcode"
)

var _ ISessionManager = (*SessionManager)(nil)

// SessionManager 管理 websocket 连接
type SessionManager struct {
	mu sync.RWMutex
	// userConns 管理一个用户多端登录的连接
	userConns  map[uint]map[string]ISession // userID -> sessionID -> ISession
	sessionMap map[string]ISession          // sessionID -> ISession
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		userConns:  make(map[uint]map[string]ISession),
		sessionMap: make(map[string]ISession),
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

func (sm *SessionManager) JoinRoom(ctx context.Context, userID uint, sessionID, roomID string) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if info, ok := sm.userConns[userID]; ok {
		if session, ok := info[sessionID]; ok {
			return session.Join(roomID)
		}
	}
	return errcode.ErrCodeSessionNotFound
}

func (sm *SessionManager) LeaveRoom(ctx context.Context, userID uint, sessionID, roomID string) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if connSet, ok := sm.userConns[userID]; ok {
		if session, ok := connSet[sessionID]; ok {
			return session.Leave(roomID)
		}
	}
	return errcode.ErrCodeSessionNotFound
}
