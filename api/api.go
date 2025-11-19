package api

import (
	"net/http"

	"github.com/miebyte/goutils/ginutils"
	middleware "github.com/superwhys/billiard-helper/api/middlewares"
	"github.com/superwhys/billiard-helper/api/router"
	"github.com/superwhys/billiard-helper/service"

	_ "github.com/superwhys/billiard-helper/cmd/swagger/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// SetupRouter godoc
// @title Billiard Helper API
// @version 1.0
// @description BilliardHelper
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func SetupRouter(services *service.Service) http.Handler {
	engine := ginutils.NewServerHandler(
		ginutils.WithMiddleware(ginutils.WithLoggingRequest(true)),
		router.AuthGroupRouter(services.AuthService),
		ginutils.WithGroupHandlers(
			ginutils.WithMiddleware(middleware.TokenVerifyMiddleware(services.AuthService)),
			ginutils.WithGroupHandlers(
				router.SocketGroupRouter(services),
				router.RoomGroupRouter(services.RoomService),
				router.ScoresGroupRouter(services.ScoresService),
			),
		),
	)

	return engine
}

func SwaggerRouter(isProd bool) http.Handler {
	if isProd {
		return http.NotFoundHandler()
	}

	return httpSwagger.Handler()
}
