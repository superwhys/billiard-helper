package service

import (
	"github.com/miebyte/goutils/redisutils"
	"github.com/miebyte/goutils/websocketutils"
	"github.com/superwhys/billiard-helper/config"
	"github.com/superwhys/billiard-helper/internal/pkg/longnet"
	"github.com/superwhys/billiard-helper/internal/ports"
)

type ServiceContext struct {
	Config      *config.Config
	RedisClient *redisutils.RedisClient
	Socket      *websocketutils.Server
	EventQueue  longnet.EventQueue
	RoomRepo    ports.RoomRepo
	PlayerRepo  ports.PlayerRepo
	ScoreRepo   ports.ScoreRepo
	UserRepo    ports.UserRepo
}
