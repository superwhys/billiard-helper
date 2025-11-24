package api

import (
	"net/http"

	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/billiard-helper/api/middlewares"
	"github.com/superwhys/billiard-helper/api/router"
	"github.com/superwhys/billiard-helper/internal/comet"
	"github.com/superwhys/billiard-helper/internal/pkg/longnet"
	"github.com/superwhys/billiard-helper/internal/service"

	_ "github.com/superwhys/billiard-helper/cmd/swagger/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type apiApp struct {
	isDev          bool
	sessionManager longnet.ISessionManager
	services       *service.Service
	cometServer    *comet.Server
}

// SetupRouter godoc
// @title Billiard Helper API
// @version 1.0
// @description BilliardHelper
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func SetupAPI(
	isDev bool,
	services *service.Service,
	cometServer *comet.Server,
) *apiApp {
	return &apiApp{
		isDev:       isDev,
		services:    services,
		cometServer: cometServer,
	}
}

func (a *apiApp) SwaggerRouter() http.Handler {
	if !a.isDev {
		return http.NotFoundHandler()
	}

	return httpSwagger.Handler()
}

func (a *apiApp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	engine := ginutils.NewServerHandler(
		ginutils.WithMiddleware(ginutils.WithLoggingRequest(true)),
		a.cometServer.Handler(),
		router.AuthGroupRouter(a.services.AuthService),
		ginutils.WithGroupHandlers(
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(a.services.AuthService)),
			ginutils.WithGroupHandlers(
				router.RoomGroupRouter(a.services.RoomService),
				router.ScoresGroupRouter(a.services.ScoresService),
			),
		),
	)

	engine.ServeHTTP(w, r)
}
