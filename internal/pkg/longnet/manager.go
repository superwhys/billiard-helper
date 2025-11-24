package longnet

import (
	"strconv"

	"github.com/miebyte/goutils/websocketutils"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type SessionManager struct {
	sessions cmap.ConcurrentMap[string, ISession]
	users    cmap.ConcurrentMap[string, []string] // userID -> []connID
}

func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		sessions: cmap.New[ISession](),
		users:    cmap.New[[]string](),
	}

	return sm
}

func (sm *SessionManager) RegisterSession(uid uint, conn websocketutils.Conn) {
	session := NewSession(uid, conn)
	sm.sessions.Set(session.ConnID(), session)

	// Update user sessions
	uidStr := strconv.FormatUint(uint64(uid), 10)
	sm.users.Upsert(uidStr, []string{session.ConnID()}, func(exist bool, valueInMap []string, newValue []string) []string {
		if !exist {
			return newValue
		}
		return append(valueInMap, newValue...)
	})
}

func (sm *SessionManager) UnregisterSession(conn websocketutils.Conn) {
	connID := conn.ID()
	session, ok := sm.sessions.Get(connID)
	if !ok {
		return
	}
	sm.sessions.Remove(connID)

	uid := session.UserID()
	uidStr := strconv.FormatUint(uint64(uid), 10)

	// Remove from user sessions
	sm.users.Upsert(uidStr, nil, func(exist bool, valueInMap []string, newValue []string) []string {
		if !exist {
			return nil
		}
		newConnIDs := make([]string, 0, len(valueInMap))
		for _, id := range valueInMap {
			if id != connID {
				newConnIDs = append(newConnIDs, id)
			}
		}
		if len(newConnIDs) == 0 {
			return newConnIDs
		}
		return newConnIDs
	})

	if v, ok := sm.users.Get(uidStr); ok && len(v) == 0 {
		sm.users.Remove(uidStr)
	}
}

func (sm *SessionManager) GetSession(connID string) ISession {
	if value, ok := sm.sessions.Get(connID); ok {
		return value
	}
	return nil
}

func (sm *SessionManager) GetSessionsByUserID(userID uint) []ISession {
	uidStr := strconv.FormatUint(uint64(userID), 10)
	if connIDs, ok := sm.users.Get(uidStr); ok {
		sessions := make([]ISession, 0, len(connIDs))
		for _, id := range connIDs {
			if s := sm.GetSession(id); s != nil {
				sessions = append(sessions, s)
			}
		}
		return sessions
	}
	return nil
}

func (sm *SessionManager) IterateSessions(callback func(ISession) bool) {
	for item := range sm.sessions.IterBuffered() {
		if !callback(item.Val) {
			break
		}
	}
}
