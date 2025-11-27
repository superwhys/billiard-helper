package user

import (
	"context"
	"time"
)

type IVerifyCodeRepository interface {
	GenerateCode(ctx context.Context, email string, ttl time.Duration) (string, error)
	GetCode(ctx context.Context, email string) (string, error)
	DeleteCode(ctx context.Context, email string) error
}

type IUserRepository interface {
	Save(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
	Update(ctx context.Context, user *User) error
}

type ISessionRepository interface {
	SetSession(ctx context.Context, userID uint, token string, ttl time.Duration) error
	GetSession(ctx context.Context, userID uint) (string, error)
	DeleteSession(ctx context.Context, userID uint) error
}
