package user

import "context"

// VerifyCodeRepository 验证码仓储接口 (基于 Redis 实现)
type IVerifyCodeRepository interface {
	// SetCode 存储验证码 (设置过期时间)
	SetCode(ctx context.Context, email string, code string, ttl int) error
	// GetCode 获取验证码
	GetCode(ctx context.Context, email string) (string, error)
	// DeleteCode 删除验证码 (验证成功后)
	DeleteCode(ctx context.Context, email string) error
}

// IUserRepository 用户仓储接口
type IUserRepository interface {
	// Create 创建用户
	Create(ctx context.Context, user *User) error
	// FindByID 根据ID查询用户
	FindByID(ctx context.Context, id uint) (*User, error)
	// FindByEmail 根据Email查询用户
	FindByEmail(ctx context.Context, email string) (*User, error)
	// Update 更新用户
	Update(ctx context.Context, user *User) error
}
