package main

import (
	"github.com/miebyte/goutils/cores"
	"github.com/miebyte/goutils/flags"
	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/mysqlutils"
	"github.com/miebyte/goutils/redisutils"
	"github.com/superwhys/billiard-helper/api"
	"github.com/superwhys/billiard-helper/config"
	"github.com/superwhys/billiard-helper/internal/comet"
	"github.com/superwhys/billiard-helper/internal/models/dbmodels"
	"github.com/superwhys/billiard-helper/internal/pkg/longnet"
	"github.com/superwhys/billiard-helper/internal/service"
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

	err = mysqlDB.AutoMigrate(dbmodels.Tables()...)
	logging.PanicError(err)

	eventQueue := longnet.NewMemoryQueue()
	sessionManager := longnet.NewSessionManager()

	services := service.NewService(config, mysqlDB, redisClient, eventQueue, sessionManager)
	cometServer := comet.NewCometServer(eventQueue, sessionManager, services)

	apiApp := api.SetupAPI(isDev(), services, cometServer)

	srv := cores.NewCores(
		cores.WithHttpCORS(),
		cores.WithRegisterService(),
		cores.WithHttpHandler("/api", apiApp),
		cores.WithHttpHandler("/swagger", apiApp.SwaggerRouter()),
		cores.WithNameWorker("CometSubscriber", cometServer.Subscribe),
	)

	logging.PanicError(cores.Start(srv, port()))
}
