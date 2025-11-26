package manager

import (
	"github.com/miebyte/goutils/websocketutils"
)

type ISession interface {
	websocketutils.Conn
	UserID() uint
	ConnID() string
	SessionID() string
}

type Session struct {
	websocketutils.Conn
	uid       uint
	sessionID string
}

func NewSession(sessionID string, uid uint, conn websocketutils.Conn) *Session {
	return &Session{
		Conn:      conn,
		uid:       uid,
		sessionID: sessionID,
	}
}

func (s *Session) UserID() uint {
	return s.uid
}

func (s *Session) ConnID() string {
	return s.ID()
}

func (s *Session) SessionID() string {
	return s.sessionID
}
