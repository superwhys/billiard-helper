package dbmodels

import (
	"github.com/superwhys/billiard-helper/internal/models/types"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email    string `gorm:"column:email;unique;type:VARCHAR(255)" json:"email"`
	Name     string `gorm:"column:name;type:VARCHAR(255)" json:"name"`
	Password string `gorm:"column:password;type:VARCHAR(255)" json:"password"`
	Avatar   string `gorm:"column:avatar;type:VARCHAR(255)" json:"avatar"`

	Rooms []*Room `json:"rooms"`
}

func (u *User) TableName() string {
	return "users"
}

func (u *User) ToType() *types.User {
	userT := &types.User{
		ID:     u.ID,
		Email:  u.Email,
		Name:   u.Name,
		Avatar: u.Avatar,
	}
	return userT
}
