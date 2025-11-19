package service

import (
	"github.com/superwhys/billiard-helper/dal/db/manager"
	"github.com/superwhys/billiard-helper/ports"
	"gorm.io/gorm"
)

type Service struct {
	ctx           *ServiceContext
	RoomService   ports.RoomService
	ScoresService ports.ScoresService
}

// NewService 创建服务集合
func NewService(db *gorm.DB) *Service {
	roomRepo := manager.NewRoomManager(db)
	playerRepo := manager.NewPlayerManager(db)
	scoreRepo := manager.NewScoresManager(db)

	ctx := &ServiceContext{
		RoomRepo:   roomRepo,
		PlayerRepo: playerRepo,
		ScoreRepo:  scoreRepo,
	}

	return &Service{
		ctx:           ctx,
		RoomService:   NewRoomService(ctx),
		ScoresService: NewScoresService(ctx),
	}
}

// Context 返回服务上下文
func (s *Service) Context() *ServiceContext {
	return s.ctx
}
