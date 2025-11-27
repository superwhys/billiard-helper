package socket

import (
	"github.com/miebyte/goutils/websocketutils"
)

type ISession interface {
	websocketutils.Conn
	UserID() uint
}

type Session struct {
	websocketutils.Conn
	userID uint
}

func NewSession(userID uint, conn websocketutils.Conn) *Session {
	return &Session{
		Conn:   conn,
		userID: userID,
	}
}

func (s *Session) UserID() uint {
	return s.userID
}

func (s *Session) ConnID() string {
	return s.ID()
}
