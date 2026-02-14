package user

type User struct {
	ID       uint     `json:"id"`
	Phone    string   `json:"phone"`
	Email    Email    `json:"email"`
	OpenID   string   `json:"open_id"`
	Name     string   `json:"name"`
	Password Password `json:"password"`
	Avatar   string   `json:"avatar"`
}

func NewUser(phone string, email Email, openID string, name string, password Password) *User {
	return &User{
		Phone:    phone,
		Email:    email,
		OpenID:   openID,
		Name:     name,
		Password: password,
	}
}

func (u *User) UpdateProfile(name string) {
	if name != "" {
		u.Name = name
	}
}
