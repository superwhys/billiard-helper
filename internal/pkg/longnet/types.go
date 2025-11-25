package longnet

import (
	"context"

	"github.com/miebyte/goutils/websocketutils"
)

type ISession interface {
	websocketutils.Conn
	UserID() uint
	ConnID() string
	SessionID() string
}

type EventQueue interface {
	Subscribe(ctx context.Context, channel string) <-chan []byte
	Publish(ctx context.Context, channel string, data *MemoryQueueMessage) error
}

type ISessionManager interface {
	RegisterSession(userID uint, sessionID string, conn websocketutils.Conn)
	UnregisterSession(conn websocketutils.Conn)
	GetUserSession(userID uint, sessionID string) ISession
	GetSessionsByUserID(userID uint) []ISession
	JoinRoom(ctx context.Context, userID uint, sessionID, roomID string) error
	LeaveRoom(ctx context.Context, userID uint, sessionID, roomID string) error
}

type MemoryQueueMessage struct {
	Event string `json:"event"`
	Data  []byte `json:"data"`
}
