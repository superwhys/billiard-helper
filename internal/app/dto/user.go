package dto

import "time"

type User struct {
	ID        uint      `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SendRegisterCodeReq struct {
	Account string `json:"account"`
}

type RegisterReq struct {
	Account  string `json:"account"`
	Code     string `json:"code"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}
