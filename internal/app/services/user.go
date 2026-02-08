package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/miebyte/goutils/logging"
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
	logging.Debugc(ctx, "verify coder sender: %s", sender.Channel())

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
func (a *UserApp) Login(ctx context.Context, req *dto.LoginReq) (*dto.TokenResponse, error) {
	if req.Password == "" && req.VerifyCode == "" {
		return nil, errcode.ErrBadRequest
	}

	// 1. 验证账号密码或者验证码
	userService := a.serviceFactory.UserService(a.repoFactory)
	u, err := userService.Login(ctx, req.Account, req.Password, req.VerifyCode)
	if err != nil {
		return nil, err
	}

	// 2. 生成 Access Token / Refresh Token
	accessToken, refreshToken, sessionID, err := jwt.GenerateTokenPair(
		[]byte(a.jwtConfig.JwtSecret),
		a.jwtConfig.JwtTimeout,
		a.jwtConfig.JwtRefreshTimeout,
		u.ID,
	)
	if err != nil {
		return nil, err
	}

	// 3. 存储 Session
	if err := a.sessionRepo.SetSession(ctx, sessionID, u.ID, a.jwtConfig.JwtRefreshTimeout); err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *UserApp) GetUserTokenClaims(ctx context.Context, tokenStr string) (*jwt.UserTokenClaims, error) {
	// 1. 解析 Token
	claims, err := jwt.ParseToken(tokenStr, []byte(a.jwtConfig.JwtSecret))
	if err != nil {
		return nil, err
	}

	if claims.TokenType != jwt.TokenTypeAccess {
		return nil, errcode.ErrUnauthorized
	}
	if claims.Subject == "" {
		return nil, errcode.ErrUnauthorized
	}

	// 2. 验证 Session (检查是否被踢出或失效)
	cachedUserID, err := a.sessionRepo.GetSession(ctx, claims.Subject)
	if err != nil {
		return nil, fmt.Errorf("session expired or invalid")
	}

	if cachedUserID != claims.UserID {
		return nil, fmt.Errorf("session user not match")
	}

	return claims, nil
}

func (a *UserApp) Logout(ctx context.Context, tokenStr string) error {
	claims, err := a.GetUserTokenClaims(ctx, tokenStr)
	if err != nil {
		return err
	}

	return a.sessionRepo.DeleteSession(ctx, claims.Subject)
}

func (a *UserApp) ForceLogoutByRefreshToken(ctx context.Context, refreshToken string) error {
	claims, err := jwt.ParseToken(refreshToken, []byte(a.jwtConfig.JwtSecret))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return errcode.ErrTokenExpired
		}
		return errcode.ErrInvalidToken
	}
	if claims.TokenType != jwt.TokenTypeRefresh {
		return errcode.ErrUnauthorized
	}
	if claims.Subject == "" {
		return errcode.ErrUnauthorized
	}
	return a.sessionRepo.DeleteSession(ctx, claims.Subject)
}

func (a *UserApp) RefreshAccessToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error) {
	claims, err := jwt.ParseToken(refreshToken, []byte(a.jwtConfig.JwtSecret))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errcode.ErrTokenExpired
		}
		return nil, errcode.ErrInvalidToken
	}
	if claims.TokenType != jwt.TokenTypeRefresh || claims.Subject == "" {
		return nil, errcode.ErrUnauthorized
	}

	cachedUserID, err := a.sessionRepo.GetSession(ctx, claims.Subject)
	if err != nil {
		return nil, errcode.ErrUnauthorized
	}
	if cachedUserID != claims.UserID {
		return nil, errcode.ErrUnauthorized
	}

	accessToken, err := jwt.GenerateAccessToken(
		[]byte(a.jwtConfig.JwtSecret),
		a.jwtConfig.JwtTimeout,
		claims.UserID,
		claims.Subject,
	)
	if err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *UserApp) GetUserInfo(ctx context.Context, userID uint) (*dto.User, error) {
	userService := a.serviceFactory.UserService(a.repoFactory)
	u, err := userService.GetUserInfo(ctx, userID)
	if err != nil {
		return nil, err
	}
	return a.userAssembler.ToDTO(u), nil
}

func (a *UserApp) UpdateSelfInfo(ctx context.Context, userID uint, req *dto.UpdateSelfInfoReq) error {
	userService := a.serviceFactory.UserService(a.repoFactory)
	return userService.UpdateProfile(ctx, userID, req.Name)
}
