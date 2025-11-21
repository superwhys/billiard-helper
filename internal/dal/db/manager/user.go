package manager

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/dal/db/query"
	"github.com/superwhys/billiard-helper/internal/models/dbmodels"
	"github.com/superwhys/billiard-helper/internal/ports"
	"gorm.io/gorm"
)

type UserRepo struct {
	query *query.Query
}

func NewUserRepo(db *gorm.DB) ports.UserRepo {
	return &UserRepo{query: query.Use(db)}
}

func (m *UserRepo) IsUserExistsByEmail(ctx context.Context, email string) (bool, error) {
	u, ud := m.query.User, m.query.User.WithContext(ctx)

	cnt, err := ud.Where(u.Email.Eq(email)).Count()
	if err != nil {
		return false, err
	}

	return cnt > 0, nil
}

func (m *UserRepo) GetUserByEmail(ctx context.Context, email string) (*dbmodels.User, error) {
	u, ud := m.query.User, m.query.User.WithContext(ctx)

	return ud.Where(u.Email.Eq(email)).First()
}

func (m *UserRepo) GetUserByID(ctx context.Context, id uint) (*dbmodels.User, error) {
	u, ud := m.query.User, m.query.User.WithContext(ctx)

	return ud.Where(u.ID.Eq(id)).First()
}

func (m *UserRepo) CreateUser(ctx context.Context, user *dbmodels.User) error {
	ud := m.query.User.WithContext(ctx)

	return ud.Create(user)
}

func (m *UserRepo) UpdateUser(ctx context.Context, user *dbmodels.User) error {
	u, ud := m.query.User, m.query.User.WithContext(ctx)

	_, err := ud.Where(u.ID.Eq(user.ID)).Omit(u.Email, u.Password).Updates(user)
	return err
}

func (m *UserRepo) DeleteUser(ctx context.Context, id uint) error {
	u, ud := m.query.User, m.query.User.WithContext(ctx)

	_, err := ud.Where(u.ID.Eq(id)).Delete()
	return err
}
