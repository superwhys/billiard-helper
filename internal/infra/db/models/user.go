package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email    string `gorm:"column:email;unique;type:VARCHAR(255)" json:"email"`
	Name     string `gorm:"column:name;type:VARCHAR(255)" json:"name"`
	Password string `gorm:"column:password;type:VARCHAR(255)" json:"password"`
	Avatar   string `gorm:"column:avatar;type:VARCHAR(255)" json:"avatar"`

	Matches []*Match `json:"matches"`
}

func (u *User) TableName() string {
	return "users"
}
