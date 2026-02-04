package main

import (
	"github.com/miebyte/goutils/cores"
	"github.com/miebyte/goutils/flags"
	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/mysqlutils"
	"github.com/miebyte/goutils/redisutils"
	"github.com/superwhys/billiard-helper/api"
	"github.com/superwhys/billiard-helper/config"
	"github.com/superwhys/billiard-helper/internal/app/factory"
	"github.com/superwhys/billiard-helper/internal/app/hook"
	"github.com/superwhys/billiard-helper/internal/app/services"
	"github.com/superwhys/billiard-helper/internal/infra/cache"
	"github.com/superwhys/billiard-helper/internal/infra/db/models"
	"github.com/superwhys/billiard-helper/internal/infra/socket"
	"github.com/superwhys/billiard-helper/internal/infra/verifycode"
	"github.com/superwhys/billiard-helper/internal/infra/verifycode/email"
	"github.com/superwhys/billiard-helper/internal/infra/verifycode/sms"
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
	smsSender := sms.NewSMSSender()
	senderFactory := verifycode.NewSenderFactory(emailSender, smsSender)

	// Initialize Repositories
	verifyCodeRepo := cache.NewVerifyCodeRepository(redisClient)
	sessionRepo := cache.NewSessionRepository(redisClient)

	repoFactory := factory.NewRepositoryFactory(mysqlDB)
	serviceFactory := factory.NewDomainServiceFactory(redisClient)

	// Initialize app services
	userApp := services.NewUserApp(
		serviceFactory,
		repoFactory,
		sessionRepo,
		verifyCodeRepo,
		senderFactory,
		config.JwtConfig,
	)
	matchApp := services.NewMatchApp(serviceFactory, repoFactory, nil, nil)
	scoreApp := services.NewScoreApp(serviceFactory, repoFactory, nil)

	socketManager := socket.NewSocketManager(hook.NewSocketHook(matchApp, config.JwtConfig))

	apiApp := api.SetupApi(isDev(), socketManager, userApp, scoreApp, matchApp)

	srv := cores.NewCores(
		cores.WithHttpCORS(),
		cores.WithRegisterService(),
		cores.WithHttpHandler("/api", apiApp),
		cores.WithHttpHandler("/swagger", apiApp.SwaggerRouter()),
	)

	logging.PanicError(cores.Start(srv, port()))
}
