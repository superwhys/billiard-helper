package services

import (
	"context"
	"fmt"

	"github.com/superwhys/billiard-helper/config"
	"github.com/superwhys/billiard-helper/internal/app/assembler"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/domain/user"
	"github.com/superwhys/billiard-helper/internal/pkg/jwt"
)

type UserApp struct {
	userService   user.IUserService
	userAssembler *assembler.UserAssembler
	sessionRepo   user.ISessionRepository
	jwtConfig     *config.JwtConfig
}

func NewUserApp(
	userService user.IUserService,
	sessionRepo user.ISessionRepository,
	jwtConfig *config.JwtConfig,
) *UserApp {
	return &UserApp{
		userService:   userService,
		userAssembler: assembler.NewUserAssembler(),
		sessionRepo:   sessionRepo,
		jwtConfig:     jwtConfig,
	}
}

// SendRegisterCode 发送注册验证码
func (a *UserApp) SendRegisterCode(ctx context.Context, req *dto.SendRegisterCodeReq) error {
	return a.userService.SendRegisterCode(ctx, req.Email)
}

// Register 注册用户
func (a *UserApp) Register(ctx context.Context, req *dto.RegisterReq) error {
	return a.userService.RegisterWithCode(ctx, req.Email, req.Code, req.Password, req.Name)
}

// Login 用户登录
func (a *UserApp) Login(ctx context.Context, req *dto.LoginReq) (string, *dto.User, error) {
	// 1. 验证账号密码
	u, err := a.userService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return "", nil, err
	}

	// 2. 生成 Access Token
	token, err := jwt.GenerateToken(
		[]byte(a.jwtConfig.JwtSecret),
		a.jwtConfig.JwtTimeout,
		u.ID,
	)
	if err != nil {
		return "", nil, err
	}

	// 3. 存储 Session
	if err := a.sessionRepo.SetSession(ctx, u.ID, token, a.jwtConfig.JwtTimeout); err != nil {
		return "", nil, err
	}

	userDTO := a.userAssembler.ToDTO(u)
	return token, userDTO, nil
}

func (a *UserApp) GetUserTokenClaims(ctx context.Context, tokenStr string) (*jwt.UserTokenClaims, error) {
	// 1. 解析 Token
	claims, err := jwt.ParseToken(tokenStr, []byte(a.jwtConfig.JwtSecret))
	if err != nil {
		return nil, err
	}

	// 2. 验证 Session (检查是否被踢出或失效)
	cachedToken, err := a.sessionRepo.GetSession(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("session expired or invalid")
	}

	if cachedToken != tokenStr {
		return nil, fmt.Errorf("account logged in on another device")
	}

	return claims, nil
}
