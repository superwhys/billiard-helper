package user

import (
	"context"
	"errors"

	"github.com/superwhys/billiard-helper/internal/errcode"
	"gorm.io/gorm"
)

type IUserService interface {
	// RegisterWithCode 使用验证码注册
	RegisterWithCode(ctx context.Context, email, password, name string) error

	// Login 处理登录校验
	Login(ctx context.Context, email, password string) (*User, error)

	// UpdateProfile 更新用户资料
	UpdateProfile(ctx context.Context, userID uint, name, avatar string) error

	// ChangePassword 修改密码
	ChangePassword(ctx context.Context, userID uint, oldPassword, newPassword string) error

	// GetUserInfo 获取用户信息
	GetUserInfo(ctx context.Context, userID uint) (*User, error)
}

var _ IUserService = (*UserService)(nil)

type UserService struct {
	userRepository IUserRepository
}

func NewUserService(userRepository IUserRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (s *UserService) RegisterWithCode(ctx context.Context, email, password, name string) error {
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

func (s *UserService) UpdateProfile(ctx context.Context, userID uint, name, avatar string) error {
	// 1. 查询用户
	user, err := s.userRepository.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// 2. 更新用户
	user.UpdateProfile(name, avatar)
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
