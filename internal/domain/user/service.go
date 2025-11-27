package user

import "context"

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
