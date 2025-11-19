package main

import (
	"github.com/miebyte/goutils/cores"
	"github.com/miebyte/goutils/flags"
	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/mysqlutils"
	"github.com/superwhys/billiard-helper/api"
	"github.com/superwhys/billiard-helper/models/config"
	"github.com/superwhys/billiard-helper/models/dbmodels"
	"github.com/superwhys/billiard-helper/service"
)

var (
	port            = flags.Int("port", 8080, "server run port")
	isDev           = flags.Bool("dev", true, "is dev")
	configFlag      = flags.Struct("config", (*config.Config)(nil), "server config")
	mysqlConfigFlag = flags.Struct("mysql", (*mysqlutils.MysqlConfig)(nil), "mysql config")
)

func main() {
	flags.Parse()

	config := new(config.Config)
	logging.PanicError(configFlag(config))

	mysqlConf := new(mysqlutils.MysqlConfig)
	logging.PanicError(mysqlConfigFlag(mysqlConf))

	mysqlDB, err := mysqlConf.DialMysqlGorm()
	logging.PanicError(err)

	err = mysqlDB.AutoMigrate(dbmodels.Tables()...)
	logging.PanicError(err)

	services := service.NewService(config, mysqlDB)
	router := api.SetupRouter(services)

	srv := cores.NewCores(
		cores.WithHttpCORS(),
		cores.WithRegisterService(),
		cores.WithHttpHandler("/api", router),
		cores.WithHttpHandler("/swagger", api.SwaggerRouter(!isDev())),
	)

	logging.PanicError(cores.Start(srv, port()))
}
