package assembler

import (
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/domain/user"
)

type UserAssembler struct{}

func NewUserAssembler() *UserAssembler {
	return &UserAssembler{}
}

func (a *UserAssembler) ToDTO(u *user.User) *dto.User {
	if u == nil {
		return nil
	}
	return &dto.User{
		ID:        u.ID,
		Email:     u.Email.String(),
		Name:      u.Name,
		Avatar:    u.Avatar,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (a *UserAssembler) ToEntity(d *dto.User) *user.User {
	if d == nil {
		return nil
	}

	email, _ := user.NewEmail(d.Email)
	return &user.User{
		ID:        d.ID,
		Email:     email,
		Name:      d.Name,
		Avatar:    d.Avatar,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		// Password: 留空或由业务逻辑处理
	}
}
