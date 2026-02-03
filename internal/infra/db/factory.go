package db

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/domain/user"
	"gorm.io/gorm"
)

type RepositoryFactory struct {
	db *gorm.DB

	MatchRepo match.IMatchRepository
	UserRepo  user.IUserRepository
}

func NewRepositoryFactory(db *gorm.DB) *RepositoryFactory {
	return &RepositoryFactory{
		db: db,

		MatchRepo: NewMatchRepo(db),
		UserRepo:  NewUserRepo(db),
	}
}

func (f *RepositoryFactory) WithTransaction(ctx context.Context, fn func(factory *RepositoryFactory) error) error {
	return f.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(NewRepositoryFactory(tx))
	})
}
