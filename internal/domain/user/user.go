package user

import "time"

type User struct {
	ID        uint
	Email     Email
	Name      string
	Password  Password
	Avatar    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(email Email, name string, password Password) *User {
	return &User{
		Email:     email,
		Name:      name,
		Password:  password,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (u *User) UpdateProfile(name, avatar string) {
	if name != "" {
		u.Name = name
	}
	if avatar != "" {
		u.Avatar = avatar
	}
	u.UpdatedAt = time.Now()
}
