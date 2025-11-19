package api

import (
	"net/http"

	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/billiard-helper/router"
	"github.com/superwhys/billiard-helper/service"
)

// SetupRouter godoc
// @title Teacher Assistant API
// @version 1.0
// @description BilliardHelper
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func SetupRouter(services *service.Service) http.Handler {
	engine := ginutils.NewServerHandler(
		ginutils.WithMiddleware(ginutils.WithLoggingRequest(true)),
		ginutils.WithHandler(http.MethodGet, "/ws", router.SocketHandler(services)),
	)

	return engine
}
