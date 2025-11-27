package user

import "time"

type User struct {
	ID        uint      `json:"id"`
	Email     Email     `json:"email"`
	Name      string    `json:"name"`
	Password  Password  `json:"password"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
