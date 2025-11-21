package longnet

import (
	"github.com/miebyte/goutils/websocketutils"
)

type Session struct {
	websocketutils.Socket
	userID uint
}

func NewSession(uid uint, conn websocketutils.Socket) *Session {
	return &Session{
		Socket: conn,
		userID: uid,
	}
}

func (s *Session) UserID() uint {
	return s.userID
}

func (s *Session) ConnID() string {
	return s.ID()
}
