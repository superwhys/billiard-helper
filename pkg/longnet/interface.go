package longnet

import "github.com/miebyte/goutils/websocketutils"

type ISession interface {
	websocketutils.Socket
	UserID() uint
	ConnID() string
}

type ISessionManager interface {
	RegisterSession(uid uint, conn websocketutils.Socket)
	GetSession(connID string) ISession
	GetSessionsByUserID(userID uint) []ISession
}
