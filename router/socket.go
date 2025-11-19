package router

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/websocketutils"
	"github.com/superwhys/billiard-helper/service"
)

func SocketHandler(services *service.Service) gin.HandlerFunc {
	socket := websocketutils.NewServer()
	setupBilliardSocket(services, socket)
	return func(ctx *gin.Context) {
		socket.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func setupBilliardSocket(services *service.Service, socket *websocketutils.Server) {
	billiardNamespace := socket.Of("billiard")
	billiardNamespace.On(websocketutils.EventConnection, func(s websocketutils.Socket) {
		s.On("join_room", func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On("leave_room", func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On("score_add", func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On("score_minus", func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On("score_reset", func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On("kick_player", func(s websocketutils.Socket, rm json.RawMessage) {})
	})
}
