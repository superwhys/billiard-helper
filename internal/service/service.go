package service

import (
	"github.com/miebyte/goutils/redisutils"
	"github.com/superwhys/billiard-helper/config"
	"github.com/superwhys/billiard-helper/internal/comet/queue"
	"github.com/superwhys/billiard-helper/internal/dal/db/manager"
	"github.com/superwhys/billiard-helper/internal/ports"
	"gorm.io/gorm"
)

type Service struct {
	ctx           *ServiceContext
	RoomService   ports.RoomService
	ScoresService ports.ScoresService
	AuthService   ports.AuthService
}

// NewService 创建服务集合
func NewService(
	config *config.Config,
	db *gorm.DB,
	redisClient *redisutils.RedisClient,
	eventQueue queue.EventQueue,
) *Service {
	roomRepo := manager.NewRoomManager(db)
	playerRepo := manager.NewPlayerManager(db)
	scoreRepo := manager.NewScoresManager(db)
	userRepo := manager.NewUserRepo(db)

	ctx := &ServiceContext{
		Config:      config,
		RedisClient: redisClient,
		RoomRepo:    roomRepo,
		PlayerRepo:  playerRepo,
		ScoreRepo:   scoreRepo,
		UserRepo:    userRepo,
		EventQueue:  eventQueue,
	}

	return &Service{
		ctx:           ctx,
		RoomService:   NewRoomService(ctx),
		ScoresService: NewScoresService(ctx),
		AuthService:   NewAuthService(ctx),
	}
}

// Context 返回服务上下文
func (s *Service) Context() *ServiceContext {
	return s.ctx
}
