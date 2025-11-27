package api

import (
	"net/http"

	"github.com/miebyte/goutils/ginutils"
	"github.com/superwhys/billiard-helper/api/middlewares"
	"github.com/superwhys/billiard-helper/api/routers"
	"github.com/superwhys/billiard-helper/internal/app/services"

	_ "github.com/superwhys/billiard-helper/cmd/swagger/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Api struct {
	isDev       bool
	userApp     *services.UserApp
	scoreApp    *services.ScoreApp
	matchApp    *services.MatchApp
	httpHandler http.Handler
}

// SetupRouter godoc
// @title Billiard Helper API
// @version 1.0
// @description BilliardHelper
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func SetupApi(
	isDev bool,
	userApp *services.UserApp,
	scoreApp *services.ScoreApp,
	matchApp *services.MatchApp,
) *Api {
	engine := ginutils.NewServerHandler(
		ginutils.WithMiddleware(ginutils.WithLoggingRequest(true)),
		// cometServer.Handler(),
		routers.AuthGroupRouter(userApp),
		ginutils.WithGroupHandlers(
			ginutils.WithMiddleware(middlewares.TokenVerifyMiddleware(userApp)),
			ginutils.WithGroupHandlers(
				routers.MatchGroupRouter(matchApp),
			// router.ScoresGroupRouter(services.ScoresService),
			),
		),
	)

	return &Api{
		isDev:       isDev,
		httpHandler: engine,
		userApp:     userApp,
		scoreApp:    scoreApp,
		matchApp:    matchApp,
	}
}

func (a *Api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.httpHandler.ServeHTTP(w, r)
}

func (a *Api) SwaggerRouter() http.Handler {
	if !a.isDev {
		return http.NotFoundHandler()
	}

	return httpSwagger.Handler()
}
