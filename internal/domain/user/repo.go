package user

import (
	"context"
	"time"
)

type IVerifyCodeRepository interface {
	GenerateCode(ctx context.Context, account string, ttl time.Duration) (string, string, error)
	GetCode(ctx context.Context, codeId, account string) (string, error)
	DeleteCode(ctx context.Context, codeId string) error
}

type IUserRepository interface {
	Save(ctx context.Context, user *User) error
	IsExists(ctx context.Context, account string) (bool, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)
	FindByOpenID(ctx context.Context, openID string) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
	List(ctx context.Context, offset, limit int) ([]*User, int64, error)
	Update(ctx context.Context, user *User) error
}

type IBlockedIdentityRepository interface {
	IsBlocked(ctx context.Context, kind BlockedIdentityKind, value string) (bool, error)
	Block(ctx context.Context, kind BlockedIdentityKind, value string) error
	Unblock(ctx context.Context, kind BlockedIdentityKind, value string) error
}

type ISessionRepository interface {
	SetSession(ctx context.Context, sessionID string, userID uint, ttl time.Duration) error
	GetSession(ctx context.Context, sessionID string) (uint, error)
	DeleteSession(ctx context.Context, sessionID string) error
}
