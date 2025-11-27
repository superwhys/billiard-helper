package user

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/superwhys/billiard-helper/internal/errcode"
	"github.com/superwhys/billiard-helper/internal/infra/email"
	"gorm.io/gorm"
)

type IUserService interface {
	// SendRegisterCode 发送注册验证码
	SendRegisterCode(ctx context.Context, email string) error

	// RegisterWithCode 使用验证码注册
	// 包含逻辑：校验验证码 -> 校验邮箱是否已存在 -> 创建用户 -> 删除验证码
	RegisterWithCode(ctx context.Context, email, code, password, name string) error

	// Login 处理登录校验
	Login(ctx context.Context, email, password string) (*User, error)

	// Logout 退出登录
	Logout(ctx context.Context, userID uint) error

	// UpdateProfile 更新用户资料
	UpdateProfile(ctx context.Context, userID uint, name, avatar string) error

	// ChangePassword 修改密码
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error

	// GetUserInfo 获取用户信息
	GetUserInfo(ctx context.Context, userID uint) (*User, error)
}

var _ IUserService = (*UserService)(nil)

type UserService struct {
	userRepository       IUserRepository
	verifyCodeRepository IVerifyCodeRepository
	emailSender          email.IEmailSender
}

func NewUserService(
	userRepository IUserRepository,
	verifyCodeRepository IVerifyCodeRepository,
	emailSender email.IEmailSender,
) *UserService {
	return &UserService{
		userRepository:       userRepository,
		verifyCodeRepository: verifyCodeRepository,
		emailSender:          emailSender,
	}
}

func (s *UserService) generateDigitCode(length int) (string, error) {
	var builder strings.Builder
	for range length {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		builder.WriteByte(byte('0' + n.Int64()))
	}
	return builder.String(), nil
}

func (s *UserService) SendRegisterCode(ctx context.Context, email string) error {
	code, err := s.generateDigitCode(6)
	if err != nil {
		return err
	}

	err = s.verifyCodeRepository.SetCode(ctx, email, code, 10*time.Minute)
	if err != nil {
		return err
	}

	return s.emailSender.SendVerifyCode(ctx, email, code)
}

func (s *UserService) RegisterWithCode(ctx context.Context, email, code, password, name string) error {
	// 1. 校验邮箱和密码
	emailObj, err := NewEmail(email)
	if err != nil {
		return err
	}

	passwordObj, err := NewPasswordFromPlain(password)
	if err != nil {
		return err
	}

	// 2. 校验邮箱是否已存在
	user, err := s.userRepository.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if user != nil {
		return errcode.ErrCodeUserAlreadyExists
	}

	// 3. 校验验证码
	storedCode, err := s.verifyCodeRepository.GetCode(ctx, email)
	if err != nil {
		return err
	}
	if code != storedCode {
		return errcode.ErrCodeInvalidCode
	}

	// 删除验证码，不需要关心是否失败
	_ = s.verifyCodeRepository.DeleteCode(ctx, email)

	userObj := NewUser(emailObj, name, passwordObj)
	return s.userRepository.Save(ctx, userObj)
}

func (s *UserService) Login(ctx context.Context, email, password string) (*User, error) {
	// 1. 查询用户
	user, err := s.userRepository.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if user == nil {
		return nil, errcode.ErrCodeUserNotFound
	}

	// 2. 校验密码
	if !user.Password.Compare(password) {
		return nil, errcode.ErrCodeInvalidPassword
	}

	return user, nil
}

func (s *UserService) Logout(ctx context.Context, userID uint) error {
	panic("not implemented")
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uint, name, avatar string) error {
	panic("not implemented")
}

func (s *UserService) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	panic("not implemented")
}

func (s *UserService) GetUserInfo(ctx context.Context, userID uint) (*User, error) {
	panic("not implemented")
}
