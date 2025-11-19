package service

import (
	"github.com/miebyte/goutils/redisutils"
	"github.com/miebyte/goutils/websocketutils"
	"github.com/superwhys/billiard-helper/models/config"
	"github.com/superwhys/billiard-helper/ports"
)

type ServiceContext struct {
	Config      *config.Config
	RedisClient *redisutils.RedisClient
	Socket      *websocketutils.Server
	RoomRepo    ports.RoomRepo
	PlayerRepo  ports.PlayerRepo
	ScoreRepo   ports.ScoreRepo
	UserRepo    ports.UserRepo
}
