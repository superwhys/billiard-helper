package request

import "github.com/superwhys/billiard-helper/internal/models/types"

type SendEmailCodeReq struct {
	Email string               `json:"email" binding:"required"`
	Scene types.EmailCodeScene `json:"scene" binding:"required"`
}

type RegisterReq struct {
	Email    string `json:"email" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginReq struct {
	Email     string          `json:"email" binding:"required"`
	LoginType types.LoginType `json:"login_type" binding:"required"`
	// if login type is code, secret is the code
	// if login type is password, secret is the password
	Secret string `json:"secret" binding:"required"`
}
