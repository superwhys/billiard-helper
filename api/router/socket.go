package router

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/miebyte/goutils/websocketutils"
	"github.com/superwhys/billiard-helper/models/constant"
	"github.com/superwhys/billiard-helper/service"
)

func SocketHandler(services *service.Service) gin.HandlerFunc {
	socket := resolveSocketServer(services)
	setupBilliardSocket(services, socket)
	return func(ctx *gin.Context) {
		socket.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func setupBilliardSocket(services *service.Service, socket *websocketutils.Server) {
	billiardNamespace := socket.Of(constant.BilliardNamespace)
	billiardNamespace.On(websocketutils.EventConnection, func(s websocketutils.Socket) {
		s.On(constant.EventJoinRoom, func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On(constant.EventLeaveRoom, func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On(constant.EventScoreAdd, func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On(constant.EventScoreMinus, func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On(constant.EventScoreReset, func(s websocketutils.Socket, rm json.RawMessage) {})

		s.On(constant.EventKickPlayer, func(s websocketutils.Socket, rm json.RawMessage) {})
	})
}

func resolveSocketServer(services *service.Service) *websocketutils.Server {
	if services != nil {
		if ctx := services.Context(); ctx != nil && ctx.Socket != nil {
			return ctx.Socket
		}
	}
	return websocketutils.NewServer()
}
