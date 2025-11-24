package longnet

import (
	"github.com/miebyte/goutils/websocketutils"
)

type Session struct {
	websocketutils.Conn
	uid  uint
	uuid string
}

func NewSession(uuid string, uid uint, conn websocketutils.Conn) *Session {
	return &Session{
		Conn: conn,
		uid:  uid,
		uuid: uuid,
	}
}

func (s *Session) UserID() uint {
	return s.uid
}

func (s *Session) ConnID() string {
	return s.ID()
}

func (s *Session) UUID() string {
	return s.uuid
}
