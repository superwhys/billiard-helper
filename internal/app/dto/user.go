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

type UpdateSelfInfoReq struct {
	Name string `json:"name"`
}

type RegisterReq struct {
	Account  string `json:"account"`
	CodeID   string `json:"code_id"`
	Code     string `json:"code"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginReq struct {
	Account    string `json:"account"`
	Password   string `json:"password"`
	CodeID     string `json:"code_id"`
	VerifyCode string `json:"verify_code"`
}

type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type WechatLoginReq struct {
	Code string `json:"code"`
}
