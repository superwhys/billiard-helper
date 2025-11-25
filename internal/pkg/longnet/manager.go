package longnet

import (
	"context"
	"strconv"

	"github.com/miebyte/goutils/websocketutils"
	cmap "github.com/orcaman/concurrent-map/v2"
	"github.com/superwhys/billiard-helper/internal/models/errcode"
	"github.com/superwhys/billiard-helper/internal/pkg/jwt"
)

type SessionManager struct {
	sessions cmap.ConcurrentMap[string, ISession]
	uuids    cmap.ConcurrentMap[string, string]   // uuid -> connID
	users    cmap.ConcurrentMap[string, []string] // userID -> []connID
}

func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		sessions: cmap.New[ISession](),
		uuids:    cmap.New[string](),
		users:    cmap.New[[]string](),
	}

	return sm
}

func (sm *SessionManager) RegisterSession(claims *jwt.UserTokenClaims, conn websocketutils.Conn) {
	session := NewSession(claims.UUID, claims.User.ID, conn)
	sm.sessions.Set(session.ConnID(), session)
	sm.uuids.Set(claims.UUID, session.ConnID())

	// Update user sessions
	uidStr := strconv.FormatUint(uint64(claims.User.ID), 10)
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

	sm.uuids.Remove(session.UUID())

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

func (sm *SessionManager) GetSessionByUUID(uuid string) ISession {
	if connID, ok := sm.uuids.Get(uuid); ok {
		return sm.GetSession(connID)
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

func (sm *SessionManager) JoinRoom(ctx context.Context, uuid, roomID string) error {
	session := sm.GetSessionByUUID(uuid)
	if session == nil {
		return errcode.ErrCodeSessionNotFound
	}

	return session.Join(roomID)
}

func (sm *SessionManager) LeaveRoom(ctx context.Context, uuid, roomID string) error {
	session := sm.GetSessionByUUID(uuid)
	if session == nil {
		return errcode.ErrCodeSessionNotFound
	}

	return session.Leave(roomID)
}
