package db

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/domain/user"
	"github.com/superwhys/billiard-helper/internal/infra/db/assembler"
	"github.com/superwhys/billiard-helper/internal/infra/db/query"
	"gorm.io/gorm"
)

var _ user.IUserRepository = (*UserRepo)(nil)

type UserRepo struct {
	query           *query.Query
	userPoAssembler *assembler.UserPoAssembler
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		query:           query.Use(db),
		userPoAssembler: assembler.NewUserPoAssembler(),
	}
}

func (r *UserRepo) Save(ctx context.Context, user *user.User) error {
	u := r.query.User

	po := r.userPoAssembler.ToPO(user)
	return u.WithContext(ctx).Create(po)
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	u := r.query.User

	po, err := u.WithContext(ctx).Where(u.Email.Eq(email)).First()
	if err != nil {
		return nil, err
	}

	return r.userPoAssembler.ToEntity(po), nil
}

func (r *UserRepo) FindByID(ctx context.Context, id uint) (*user.User, error) {
	u := r.query.User

	po, err := u.WithContext(ctx).Where(u.ID.Eq(id)).First()
	if err != nil {
		return nil, err
	}

	return r.userPoAssembler.ToEntity(po), nil
}

func (r *UserRepo) Update(ctx context.Context, user *user.User) error {
	u := r.query.User

	po := r.userPoAssembler.ToPO(user)

	resp, err := u.WithContext(ctx).
		Where(u.ID.Eq(user.ID)).
		Select(u.Name, u.Password, u.Avatar, u.UpdatedAt).
		Updates(po)
	if err != nil {
		return err
	}

	if resp.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
