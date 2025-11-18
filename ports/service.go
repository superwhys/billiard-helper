package ports

import (
	"context"
	"net/http"
)

type SocketService interface {
	http.Handler
	JoinRoom(ctx context.Context, room string) error
}

type RoomService interface{}
