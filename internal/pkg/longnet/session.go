package longnet

import (
	"github.com/miebyte/goutils/websocketutils"
)

type Session struct {
	websocketutils.Conn
	uid uint
}

func NewSession(uid uint, conn websocketutils.Conn) *Session {
	return &Session{
		Conn: conn,
		uid:  uid,
	}
}

func (s *Session) UserID() uint {
	return s.uid
}

func (s *Session) ConnID() string {
	return s.ID()
}
