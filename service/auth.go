package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/miebyte/goutils/emailutils"
	"github.com/miebyte/goutils/logging"
	"github.com/redis/go-redis/v9"
	"github.com/superwhys/billiard-helper/dal/cache"
	"github.com/superwhys/billiard-helper/models/dbmodels"
	"github.com/superwhys/billiard-helper/models/errcode"
	"github.com/superwhys/billiard-helper/models/request"
	"github.com/superwhys/billiard-helper/models/types"
	"github.com/superwhys/billiard-helper/pkg/emailcode"
	"github.com/superwhys/billiard-helper/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	emailCodeLength = 6
	emailSubject    = "Billiard Helper - 邮箱验证码"
	emailBody       = "你的验证码是: %s"
)

type AuthService struct {
	srvCtx      *ServiceContext
	emailClient *emailutils.EmailClient
	signingKey  []byte
}

func NewAuthService(ctx *ServiceContext) *AuthService {
	emailClient := emailutils.NewEmailClient(ctx.Config.EmailConfig)
	return &AuthService{
		srvCtx:      ctx,
		emailClient: emailClient,
		signingKey:  []byte(ctx.Config.JwtConfig.JwtSecret),
	}
}

func (s *AuthService) SendEmailCode(ctx context.Context, req *request.SendEmailCodeReq) error {
	if req == nil || req.Email == "" {
		return errcode.ErrCodeInvalidRequest
	}

	email := emailcode.NormalizeEmail(req.Email)
	scene := req.Scene
	if scene != types.EmailCodeSceneRegister && scene != types.EmailCodeSceneLogin {
		return errcode.ErrCodeInvalidRequest
	}
	emailKey := emailcode.BuildEmailCodeKey(scene, email)

	client := s.srvCtx.RedisClient
	cooldownCache := cache.EmailCodeCooldownCache(emailKey)
	codeCache := cache.EmailCodeCache(emailKey)

	if _, err := cooldownCache.Get(ctx, client); err == nil {
		return errcode.ErrCodeEmailCodeCooldown
	} else if !errors.Is(err, redis.Nil) {
		logging.Errorc(ctx, "get email cooldown failed: %v", err)
		return errcode.ErrCodeSendEmailCodeFailed
	}

	code, err := emailcode.GenerateDigitCode(emailCodeLength)
	if err != nil {
		logging.Errorc(ctx, "generate email code failed: %v", err)
		return errcode.ErrCodeSendEmailCodeFailed
	}

	if err = codeCache.Set(ctx, client, code); err != nil {
		logging.Errorc(ctx, "set email code failed: %v", err)
		return errcode.ErrCodeSendEmailCodeFailed
	}

	if err = cooldownCache.Set(ctx, client, "1"); err != nil {
		logging.Errorc(ctx, "set cooldown failed: %v", err)
	}

	if err = s.emailClient.Send(email, emailSubject, fmt.Sprintf(emailBody, code)); err != nil {
		logging.Errorc(ctx, "send email code failed: %v", err)
		return errcode.ErrCodeSendEmailCodeFailed
	}

	logging.Infof("send email code success, scene=%s email=%s code=%s", scene, email, code)
	return nil
}

func (s *AuthService) Register(ctx context.Context, req *request.RegisterReq) error {
	if req == nil {
		return errcode.ErrCodeInvalidRequest
	}

	rdb := s.srvCtx.RedisClient

	email := emailcode.NormalizeEmail(req.Email)
	code, err := emailcode.GetEmailCode(ctx, rdb, types.EmailCodeSceneRegister, email)
	if err != nil {
		return err
	}
	if code != req.Code {
		return errcode.ErrCodeEmailCodeInvalid
	}

	user, err := s.srvCtx.UserRepo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if user != nil {
		return errcode.ErrCodeUserAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	newUser := &dbmodels.User{
		Email:    email,
		Name:     req.Name,
		Password: string(passwordHash),
	}

	if err = s.srvCtx.UserRepo.CreateUser(ctx, newUser); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return errcode.ErrCodeUserAlreadyExists
		}
		return err
	}

	if err = emailcode.DeleteEmailCode(ctx, rdb, types.EmailCodeSceneRegister, email); err != nil {
		logging.Errorc(ctx, "delete email code failed: %v", err)
	}

	return nil
}

func (s *AuthService) Login(ctx context.Context, req *request.LoginReq) (string, error) {
	if req == nil {
		return "", errcode.ErrCodeInvalidRequest
	}
	email := emailcode.NormalizeEmail(req.Email)

	user, err := s.srvCtx.UserRepo.GetUserByEmail(ctx, email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	if user == nil {
		return "", errcode.ErrCodeUserNotFound
	}

	rdb := s.srvCtx.RedisClient

	switch req.LoginType {
	case types.LoginTypePassword:
		if req.Secret == "" {
			return "", errcode.ErrCodeInvalidRequest
		}
		if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Secret)); err != nil {
			return "", errcode.ErrCodeInvalidRequest
		}
	case types.LoginTypeCode:
		code, codeErr := emailcode.GetEmailCode(ctx, rdb, types.EmailCodeSceneLogin, email)
		if codeErr != nil {
			return "", codeErr
		}
		if code != req.Secret {
			return "", errcode.ErrCodeEmailCodeInvalid
		}
		if err = emailcode.DeleteEmailCode(ctx, rdb, types.EmailCodeSceneLogin, email); err != nil {
			logging.Errorc(ctx, "delete email code after login failed: %v", err)
		}
	default:
		return "", errcode.ErrCodeInvalidRequest
	}

	token, err := jwt.GenerateToken(s.signingKey, s.srvCtx.Config.JwtConfig.JwtTimeout, user.ToType())
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) GetUserTokenClaims(ctx context.Context, tokenStr string) (*jwt.UserTokenClaims, error) {
	return jwt.ParseToken(tokenStr, s.signingKey)
}
