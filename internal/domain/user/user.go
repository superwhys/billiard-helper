package user

type User struct {
	ID       uint     `json:"id"`
	Email    Email    `json:"email"`
	Name     string   `json:"name"`
	Password Password `json:"password"`
	Avatar   string   `json:"avatar"`
}

func NewUser(email Email, name string, password Password) *User {
	return &User{
		Email:    email,
		Name:     name,
		Password: password,
	}
}

func (u *User) UpdateProfile(name, avatar string) {
	if name != "" {
		u.Name = name
	}
	if avatar != "" {
		u.Avatar = avatar
	}
}
