package user

import (
	"context"
	"errors"

	"github.com/superwhys/billiard-helper/internal/errcode"
	"github.com/superwhys/billiard-helper/internal/pkg/account"
	"gorm.io/gorm"
)

type IUserService interface {
	// RegisterUser 注册用户
	RegisterUser(ctx context.Context, account, password, name string) error

	// Login 处理登录校验
	Login(ctx context.Context, account, password, verifyCode string) (*User, error)

	// UpdateProfile 更新用户资料
	UpdateProfile(ctx context.Context, userID uint, name string) error

	// ChangePassword 修改密码
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error

	// GetUserInfo 获取用户信息
	GetUserInfo(ctx context.Context, userID uint) (*User, error)
}

var _ IUserService = (*UserService)(nil)

type UserService struct {
	userRepository IUserRepository
	verifyCodeRepo IVerifyCodeRepository
}

func NewUserService(userRepository IUserRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (s *UserService) RegisterUser(ctx context.Context, acc, password, name string) (err error) {
	user := &User{
		Name: name,
	}

	switch {
	case account.IsEmailAccount(acc):
		user.Email, err = NewEmail(acc)
		if err != nil {
			return err
		}
	case account.IsPhoneAccount(acc):
		user.Phone = acc
	default:
		return errcode.ErrBadRequest
	}

	user.Password, err = NewPasswordFromPlain(password)
	if err != nil {
		return err
	}

	exists, err := s.userRepository.IsExists(ctx, acc)
	if err != nil {
		return err
	}
	if exists {
		return errcode.ErrCodeUserAlreadyExists
	}

	return s.userRepository.Save(ctx, user)
}

func (s *UserService) Login(ctx context.Context, acc, password, verifyCode string) (*User, error) {
	// 1. 查询用户
	var (
		user *User
		err  error
	)
	switch {
	case account.IsEmailAccount(acc):
		user, err = s.userRepository.FindByEmail(ctx, acc)
	case account.IsPhoneAccount(acc):
		user, err = s.userRepository.FindByPhone(ctx, acc)
	default:
		return nil, errcode.ErrBadRequest
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if user == nil {
		return nil, errcode.ErrCodeUserNotFound
	}

	// 2. 校验密码/验证码
	switch {
	case password != "":
		if !user.Password.Compare(password) {
			return nil, errcode.ErrCodeInvalidPassword
		}
	case verifyCode != "":
		storedCode, err := s.verifyCodeRepo.GetCode(ctx, acc)
		if err != nil {
			return nil, err
		}
		if verifyCode != storedCode {
			return nil, errcode.ErrCodeInvalidCode
		}
		_ = s.verifyCodeRepo.DeleteCode(ctx, acc)
	default:
		return nil, errcode.ErrBadRequest
	}

	return user, nil
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uint, name string) error {
	// 1. 查询用户
	user, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// 2. 更新用户
	user.UpdateProfile(name)
	return s.userRepository.Update(ctx, user)
}

func (s *UserService) ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error {
	// 1. 查询用户
	user, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// 2. 校验旧密码
	if !user.Password.Compare(oldPassword) {
		return errcode.ErrCodeInvalidPassword
	}

	// 3. 更新密码
	user.Password, err = NewPasswordFromPlain(newPassword)
	if err != nil {
		return err
	}

	// 4. 更新用户
	return s.userRepository.Update(ctx, user)
}

func (s *UserService) GetUserInfo(ctx context.Context, userID uint) (*User, error) {
	user, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
