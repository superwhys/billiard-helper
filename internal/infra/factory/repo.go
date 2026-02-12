package factory

import (
	"context"

	appfactory "github.com/superwhys/billiard-helper/internal/app/factory"
	"github.com/superwhys/billiard-helper/internal/domain/event"
	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/domain/user"
	"github.com/superwhys/billiard-helper/internal/infra/db"
	"gorm.io/gorm"
)

type repositoryFactory struct {
	db *gorm.DB

	matchRepo     match.IMatchRepository
	userRepo      user.IUserRepository
	playerRepo    match.IPlayerRepository
	matchGameRepo match.IMatchGameRepository
	eventRepo     event.IEventRepository
}

// NewRepositoryFactory 创建仓储工厂
func NewRepositoryFactory(gormDB *gorm.DB) appfactory.IRepoFactory {
	return &repositoryFactory{
		db: gormDB,

		matchRepo:     db.NewMatchRepo(gormDB),
		userRepo:      db.NewUserRepo(gormDB),
		playerRepo:    db.NewPlayerRepo(gormDB),
		matchGameRepo: db.NewMatchGameRepo(gormDB),
		eventRepo:     db.NewMatchEventRepo(gormDB),
	}
}

func (f *repositoryFactory) WithTransaction(ctx context.Context, fn func(factory appfactory.IRepoFactory) error) error {
	return f.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(NewRepositoryFactory(tx))
	})
}

func (f *repositoryFactory) UserRepo() user.IUserRepository {
	return f.userRepo
}

func (f *repositoryFactory) MatchRepo() match.IMatchRepository {
	return f.matchRepo
}

func (f *repositoryFactory) PlayerRepo() match.IPlayerRepository {
	return f.playerRepo
}

func (f *repositoryFactory) MatchGameRepo() match.IMatchGameRepository {
	return f.matchGameRepo
}

func (f *repositoryFactory) EventRepo() event.IEventRepository {
	return f.eventRepo
}
