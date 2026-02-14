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

	emailStr := ""
	if po.Email != nil {
		emailStr = *po.Email
	}
	phone := ""
	if po.Phone != nil {
		phone = *po.Phone
	}
	openID := ""
	if po.OpenID != nil {
		openID = *po.OpenID
	}

	email, _ := user.NewEmail(emailStr)
	password := user.NewPasswordFromHash(po.Password)
	return &user.User{
		ID:       po.ID,
		Phone:    phone,
		Email:    email,
		OpenID:   openID,
		Name:     po.Name,
		Password: password,
		Avatar:   po.Avatar,
	}
}

func (a *UserPoAssembler) ToPO(entity *user.User) *models.User {
	if entity == nil {
		return nil
	}

	emailStr := entity.Email.String()
	var email *string
	if emailStr != "" {
		email = &emailStr
	}

	phoneStr := entity.Phone
	var phone *string
	if phoneStr != "" {
		phone = &phoneStr
	}

	openIDStr := entity.OpenID
	var openID *string
	if openIDStr != "" {
		openID = &openIDStr
	}

	return &models.User{
		Model: gorm.Model{
			ID: entity.ID,
		},
		Phone:    phone,
		Email:    email,
		OpenID:   openID,
		Name:     entity.Name,
		Password: entity.Password.Hash(),
		Avatar:   entity.Avatar,
	}
}
