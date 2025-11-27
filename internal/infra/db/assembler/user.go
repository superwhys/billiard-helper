package assembler

import (
	"github.com/superwhys/billiard-helper/internal/domain/user"
	"github.com/superwhys/billiard-helper/internal/infra/db/models"
	"gorm.io/gorm"
)

type UserPoAssembler struct{}

func NewUserPoAssembler() *UserPoAssembler {
	return &UserPoAssembler{}
}

func (a *UserPoAssembler) ToEntity(po *models.User) *user.User {
	if po == nil {
		return nil
	}

	email, _ := user.NewEmail(po.Email)
	password := user.NewPasswordFromHash(po.Password)
	return &user.User{
		ID:        po.ID,
		Email:     email,
		Name:      po.Name,
		Password:  password,
		Avatar:    po.Avatar,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}

func (a *UserPoAssembler) ToPO(entity *user.User) *models.User {
	if entity == nil {
		return nil
	}

	return &models.User{
		Model: gorm.Model{
			ID: entity.ID,
		},
		Email:    entity.Email.String(),
		Name:     entity.Name,
		Password: entity.Password.Hash(),
		Avatar:   entity.Avatar,
	}
}
