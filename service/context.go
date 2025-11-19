package service

import (
	"github.com/miebyte/goutils/websocketutils"
	"github.com/superwhys/billiard-helper/models/config"
	"github.com/superwhys/billiard-helper/ports"
)

type ServiceContext struct {
	Config *config.Config
	Socket *websocketutils.Server

	RoomRepo   ports.RoomRepo
	PlayerRepo ports.PlayerRepo
	ScoreRepo  ports.ScoreRepo
}
