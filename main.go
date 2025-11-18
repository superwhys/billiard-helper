package main

import (
	"github.com/miebyte/goutils/cores"
	"github.com/miebyte/goutils/flags"
	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/billiard-helper/api"
	"github.com/superwhys/billiard-helper/service"
)

var (
	port = flags.Int("port", 8080, "server run port")
)

func main() {
	flags.Parse()

	services := service.NewService()
	router := api.SetupRouter(services)

	srv := cores.NewCores(
		cores.WithHttpCORS(),
		cores.WithRegisterService(),
		cores.WithHttpHandler("/api", router),
	)

	logging.PanicError(cores.Start(srv, port()))
}
