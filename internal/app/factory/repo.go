package factory

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/domain/user"
	"github.com/superwhys/billiard-helper/internal/infra/db"
	"gorm.io/gorm"
)

type IRepoFactory interface {
	UserRepo() user.IUserRepository
	MatchRepo() match.IMatchRepository
	PlayerRepo() match.PlayerRepository
}

type repositoryFactory struct {
	db *gorm.DB

	matchRepo  match.IMatchRepository
	userRepo   user.IUserRepository
	playerRepo match.PlayerRepository
}

func NewRepositoryFactory(gormDB *gorm.DB) *repositoryFactory {
	return &repositoryFactory{
		db: gormDB,

		matchRepo:  db.NewMatchRepo(gormDB),
		userRepo:   db.NewUserRepo(gormDB),
		playerRepo: nil,
	}
}

func (f *repositoryFactory) WithTransaction(ctx context.Context, fn func(factory IRepoFactory) error) error {
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

func (f *repositoryFactory) PlayerRepo() match.PlayerRepository {
	return f.playerRepo
}
