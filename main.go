package main

import (
	"github.com/miebyte/goutils/cores"
	"github.com/miebyte/goutils/flags"
	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/mysqlutils"
	"github.com/superwhys/billiard-helper/api"
	"github.com/superwhys/billiard-helper/models/dbmodels"
	"github.com/superwhys/billiard-helper/service"
)

var (
	port            = flags.Int("port", 8080, "server run port")
	mysqlConfigFlag = flags.Struct("mysql", (*mysqlutils.MysqlConfig)(nil), "mysql config")
)

func main() {
	flags.Parse()

	mysqlConf := new(mysqlutils.MysqlConfig)
	logging.PanicError(mysqlConfigFlag(mysqlConf))

	mysqlDB, err := mysqlConf.DialMysqlGorm()
	logging.PanicError(err)

	err = mysqlDB.AutoMigrate(dbmodels.Tables()...)
	logging.PanicError(err)

	services := service.NewService()
	router := api.SetupRouter(services)

	srv := cores.NewCores(
		cores.WithHttpCORS(),
		cores.WithRegisterService(),
		cores.WithHttpHandler("/api", router),
	)

	logging.PanicError(cores.Start(srv, port()))
}
