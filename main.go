package main

import (
	"github.com/miebyte/goutils/cores"
	"github.com/miebyte/goutils/flags"
	"github.com/miebyte/goutils/logging"
)

var (
	port = flags.Int("port", 8080, "server run port")
)

func main() {
	flags.Parse()

	srv := cores.NewCores()

	logging.PanicError(cores.Start(srv, port()))
}
