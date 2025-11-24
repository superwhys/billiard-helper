package longnet

import (
	"github.com/miebyte/goutils/websocketutils"
)

type Session struct {
	websocketutils.Conn
	userID uint
}

func NewSession(uid uint, conn websocketutils.Conn) *Session {
	return &Session{
		Conn:   conn,
		userID: uid,
	}
}

func (s *Session) UserID() uint {
	return s.userID
}

func (s *Session) ConnID() string {
	return s.ID()
}
