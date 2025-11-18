package service

import (
	"github.com/miebyte/goutils/websocketutils"
	"github.com/superwhys/billiard-helper/models/config"
)

type ServiceContext struct {
	Config *config.Config
	Socket *websocketutils.Server
}
