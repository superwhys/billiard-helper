package db

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/domain/user"
	"github.com/superwhys/billiard-helper/internal/infra/db/models"
	"github.com/superwhys/billiard-helper/internal/infra/db/query"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ user.IBlockedIdentityRepository = (*BlockedIdentityRepo)(nil)

type BlockedIdentityRepo struct {
	query *query.Query
}

func NewBlockedIdentityRepo(db *gorm.DB) *BlockedIdentityRepo {
	return &BlockedIdentityRepo{query: query.Use(db)}
}

func (r *BlockedIdentityRepo) IsBlocked(ctx context.Context, kind user.BlockedIdentityKind, value string) (bool, error) {
	value, err := user.NormalizeBlockedIdentity(kind, value)
	if err != nil {
		return false, err
	}

	b := r.query.BlockedIdentity
	count, err := b.WithContext(ctx).Where(b.Kind.Eq(string(kind)), b.Value.Eq(value)).Count()
	return count > 0, err
}

func (r *BlockedIdentityRepo) Block(ctx context.Context, kind user.BlockedIdentityKind, value string) error {
	value, err := user.NormalizeBlockedIdentity(kind, value)
	if err != nil {
		return err
	}

	return r.query.BlockedIdentity.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&models.BlockedIdentity{Kind: string(kind), Value: value})
}

func (r *BlockedIdentityRepo) Unblock(ctx context.Context, kind user.BlockedIdentityKind, value string) error {
	value, err := user.NormalizeBlockedIdentity(kind, value)
	if err != nil {
		return err
	}

	b := r.query.BlockedIdentity
	_, err = b.WithContext(ctx).Where(b.Kind.Eq(string(kind)), b.Value.Eq(value)).Delete()
	return err
}
