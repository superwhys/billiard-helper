package longnet

import (
	"github.com/miebyte/goutils/websocketutils"
)

type ISession interface {
	websocketutils.Socket
	UserID() uint
	ConnID() string
}

type EventQueue interface {
	Subscribe(channel string, callback func(data []byte))
	Publish(channel string, data []byte)
}

type ISessionManager interface {
	RegisterSession(uid uint, conn websocketutils.Socket)
	UnregisterSession(conn websocketutils.Socket)
	GetSession(connID string) ISession
	GetSessionsByUserID(userID uint) []ISession
	IterateSessions(callback func(ISession) bool)
}
