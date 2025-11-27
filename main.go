package main

import (
	"github.com/miebyte/goutils/cores"
	"github.com/miebyte/goutils/flags"
	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/mysqlutils"
	"github.com/miebyte/goutils/redisutils"
	"github.com/superwhys/billiard-helper/api"
	"github.com/superwhys/billiard-helper/config"
	"github.com/superwhys/billiard-helper/internal/app/services"
	"github.com/superwhys/billiard-helper/internal/domain/user"
	"github.com/superwhys/billiard-helper/internal/infra/cache"
	"github.com/superwhys/billiard-helper/internal/infra/db"
	"github.com/superwhys/billiard-helper/internal/infra/db/models"
	"github.com/superwhys/billiard-helper/internal/infra/email"
)

var (
	port            = flags.Int("port", 8080, "server run port")
	isDev           = flags.Bool("dev", true, "is dev")
	configFlag      = flags.Struct("config", (*config.Config)(nil), "server config")
	redisConfigFlag = flags.Struct("redis", (*redisutils.RedisConfig)(nil), "redis config")
	mysqlConfigFlag = flags.Struct("mysql", (*mysqlutils.MysqlConfig)(nil), "mysql config")
)

func main() {
	flags.Parse()

	config := new(config.Config)
	logging.PanicError(configFlag(config))

	redisConf := new(redisutils.RedisConfig)
	logging.PanicError(redisConfigFlag(redisConf))

	redisClient, err := redisConf.DialGORedisClient()
	logging.PanicError(err)

	mysqlConf := new(mysqlutils.MysqlConfig)
	logging.PanicError(mysqlConfigFlag(mysqlConf))

	mysqlDB, err := mysqlConf.DialMysqlGorm()
	logging.PanicError(err)

	err = mysqlDB.AutoMigrate(models.Tables()...)
	logging.PanicError(err)

	emailSender := email.NewEmailSender(config.EmailConfig)

	// Initialize Repositories
	verifyCodeRepo := cache.NewVerifyCodeRepository(redisClient)
	sessionRepo := cache.NewSessionRepository(redisClient)
	userRepo := db.NewUserRepo(mysqlDB)

	// Initialize domain services
	userService := user.NewUserService(userRepo, verifyCodeRepo, emailSender)

	// Initialize app services
	userApp := services.NewUserApp(userService, sessionRepo, config.JwtConfig)
	scoreApp := services.NewScoreApp(nil, nil, nil, nil)
	matchApp := services.NewMatchApp(nil, nil, nil, nil)

	apiApp := api.SetupApi(isDev(), userApp, scoreApp, matchApp)

	srv := cores.NewCores(
		cores.WithHttpCORS(),
		cores.WithRegisterService(),
		cores.WithHttpHandler("/api", apiApp),
		cores.WithHttpHandler("/swagger", apiApp.SwaggerRouter()),
	)

	logging.PanicError(cores.Start(srv, port()))
}
