package longnet

import (
	"context"

	"github.com/miebyte/goutils/websocketutils"
	"github.com/superwhys/billiard-helper/internal/pkg/jwt"
)

type ISession interface {
	websocketutils.Conn
	UserID() uint
	ConnID() string
}

type EventQueue interface {
	Subscribe(ctx context.Context, channel string) <-chan []byte
	Publish(ctx context.Context, channel string, data *MemoryQueueMessage) error
}

type ISessionManager interface {
	RegisterSession(claims *jwt.UserTokenClaims, conn websocketutils.Conn)
	UnregisterSession(conn websocketutils.Conn)
	GetSession(connID string) ISession
	GetSessionsByUserID(userID uint) []ISession
	IterateSessions(callback func(ISession) bool)
}

type MemoryQueueMessage struct {
	Event string `json:"event"`
	Data  []byte `json:"data"`
}
