package user

import (
	"context"
	"time"
)

type IVerifyCodeRepository interface {
	GenerateCode(ctx context.Context, account string, ttl time.Duration) (string, error)
	GetCode(ctx context.Context, account string) (string, error)
	DeleteCode(ctx context.Context, account string) error
}

type IUserRepository interface {
	Save(ctx context.Context, user *User) error
	IsExists(ctx context.Context, account string) (bool, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
	Update(ctx context.Context, user *User) error
}

type ISessionRepository interface {
	SetSession(ctx context.Context, userID uint, token string, ttl time.Duration) error
	GetSession(ctx context.Context, userID uint) (string, error)
	DeleteSession(ctx context.Context, userID uint) error
}
