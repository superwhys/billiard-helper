package services

import (
	"context"
	"fmt"
	"time"

	"github.com/superwhys/billiard-helper/config"
	"github.com/superwhys/billiard-helper/internal/app/assembler"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/app/factory"
	"github.com/superwhys/billiard-helper/internal/domain/user"
	"github.com/superwhys/billiard-helper/internal/errcode"
	"github.com/superwhys/billiard-helper/internal/infra/verifycode"
	"github.com/superwhys/billiard-helper/internal/pkg/jwt"
)

type UserApp struct {
	repoFactory             factory.IRepoFactory
	serviceFactory          *factory.DomainServiceFactory
	userAssembler           *assembler.UserAssembler
	sessionRepo             user.ISessionRepository
	verifyCodeRepo          user.IVerifyCodeRepository
	verifyCodeSenderFactory verifycode.SenderFactory
	jwtConfig               *config.JwtConfig
}

func NewUserApp(
	serviceFactory *factory.DomainServiceFactory,
	repoFactory factory.IRepoFactory,
	sessionRepo user.ISessionRepository,
	verifyCodeRepo user.IVerifyCodeRepository,
	senderFactory verifycode.SenderFactory,
	jwtConfig *config.JwtConfig,
) *UserApp {
	return &UserApp{
		serviceFactory:          serviceFactory,
		repoFactory:             repoFactory,
		userAssembler:           assembler.NewUserAssembler(),
		sessionRepo:             sessionRepo,
		verifyCodeRepo:          verifyCodeRepo,
		verifyCodeSenderFactory: senderFactory,
		jwtConfig:               jwtConfig,
	}
}

// SendRegisterCode 发送注册验证码
func (a *UserApp) SendRegisterCode(ctx context.Context, req *dto.SendRegisterCodeReq) error {
	code, err := a.verifyCodeRepo.GenerateCode(ctx, req.Account, time.Minute*10)
	if err != nil {
		return err
	}

	sender, err := a.verifyCodeSenderFactory.Pick(req.Account)
	if err != nil {
		return err
	}

	return sender.SendVerifyCode(ctx, req.Account, code)
}

// Register 注册用户
func (a *UserApp) Register(ctx context.Context, req *dto.RegisterReq) error {
	// 1. 校验验证码
	storedCode, err := a.verifyCodeRepo.GetCode(ctx, req.Account)
	if err != nil {
		return err
	}
	if req.Code != storedCode {
		return errcode.ErrCodeInvalidCode
	}

	// 删除验证码，不需要关心是否失败
	_ = a.verifyCodeRepo.DeleteCode(ctx, req.Account)

	userService := a.serviceFactory.UserService(a.repoFactory)
	return userService.RegisterUser(ctx, req.Account, req.Password, req.Name)
}

// Login 用户登录
func (a *UserApp) Login(ctx context.Context, req *dto.LoginReq) (string, *dto.User, error) {
	if req.Password == "" && req.VerifyCode == "" {
		return "", nil, errcode.ErrBadRequest
	}

	// 1. 验证账号密码或者验证码
	userService := a.serviceFactory.UserService(a.repoFactory)
	u, err := userService.Login(ctx, req.Account, req.Password, req.VerifyCode)
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

func (a *UserApp) Logout(ctx context.Context, tokenStr string) error {
	claims, err := a.GetUserTokenClaims(ctx, tokenStr)
	if err != nil {
		return err
	}

	return a.sessionRepo.DeleteSession(ctx, claims.UserID)
}
