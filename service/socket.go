package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/miebyte/goutils/websocketutils"
)

type SocketService struct {
	srvCtx *ServiceContext
	socket *websocketutils.Server
}

func NewSocketService(srvCtx *ServiceContext) *SocketService {
	socket := websocketutils.NewServer()
	s := &SocketService{
		srvCtx: srvCtx,
		socket: socket,
	}

	s.initBilliardSocket()

	return s
}

func (s *SocketService) initBilliardSocket() {
	s.socket.Of("billiard").On(websocketutils.EventConnection, func(s websocketutils.Socket) {
		s.On("join", func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On("leave", func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On("score_add", func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On("score_minus", func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On("score_reset", func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On("kick_player", func(s websocketutils.Socket, rm json.RawMessage) {})
	})

}

func (s *SocketService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.srvCtx.Socket.ServeHTTP(w, r)
}

func (s *SocketService) JoinRoom(ctx context.Context, room string) error {
	return nil
}
