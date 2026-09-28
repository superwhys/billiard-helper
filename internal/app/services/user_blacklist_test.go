package services

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/superwhys/billiard-helper/config"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/app/factory"
	"github.com/superwhys/billiard-helper/internal/domain/user"
	"github.com/superwhys/billiard-helper/internal/errcode"
	"github.com/superwhys/billiard-helper/internal/pkg/jwt"
)

type blacklistTestFactory struct {
	factory.IRepoFactory
	userRepo    user.IUserRepository
	blockedRepo user.IBlockedIdentityRepository
}

func (f blacklistTestFactory) UserRepo() user.IUserRepository { return f.userRepo }
func (f blacklistTestFactory) BlockedIdentityRepo() user.IBlockedIdentityRepository {
	return f.blockedRepo
}

type blacklistTestUserRepo struct {
	user.IUserRepository
	current *user.User
}

func (r blacklistTestUserRepo) FindByID(context.Context, uint) (*user.User, error) {
	return r.current, nil
}

func (r blacklistTestUserRepo) FindByPhone(context.Context, string) (*user.User, error) {
	return r.current, nil
}

type blacklistTestBlockedRepo struct {
	user.IBlockedIdentityRepository
	kind  user.BlockedIdentityKind
	value string
}

func (r blacklistTestBlockedRepo) IsBlocked(_ context.Context, kind user.BlockedIdentityKind, value string) (bool, error) {
	normalized, err := user.NormalizeBlockedIdentity(kind, value)
	if err != nil {
		return false, err
	}
	return kind == r.kind && normalized == r.value, nil
}

type blacklistTestSessionRepo struct {
	user.ISessionRepository
	userID uint
}

type blacklistTestTransport func(*http.Request) (*http.Response, error)

func (f blacklistTestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func (r blacklistTestSessionRepo) GetSession(context.Context, string) (uint, error) {
	return r.userID, nil
}

func TestBlacklistedUserCannotUseExistingTokens(t *testing.T) {
	ctx := context.Background()
	email, err := user.NewEmail("Blocked@Example.com")
	if err != nil {
		t.Fatal(err)
	}
	const userID = 42
	app := &UserApp{
		repoFactory: blacklistTestFactory{
			userRepo:    blacklistTestUserRepo{current: &user.User{ID: userID, Email: email}},
			blockedRepo: blacklistTestBlockedRepo{kind: user.KindEmail, value: "blocked@example.com"},
		},
		sessionRepo: blacklistTestSessionRepo{userID: userID},
		jwtConfig:   &config.JwtConfig{JwtSecret: "blacklist-test-secret"},
	}
	access, refresh, _, err := jwt.GenerateTokenPair([]byte(app.jwtConfig.JwtSecret), time.Hour, time.Hour, userID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.GetUserTokenClaims(ctx, access); !errors.Is(err, errcode.ErrForbidden) {
		t.Fatalf("access token: expected forbidden, got %v", err)
	}
	if _, err := app.RefreshAccessToken(ctx, refresh); !errors.Is(err, errcode.ErrForbidden) {
		t.Fatalf("refresh token: expected forbidden, got %v", err)
	}
}

func TestBlacklistedEmailCannotRequestCode(t *testing.T) {
	app := &UserApp{repoFactory: blacklistTestFactory{
		blockedRepo: blacklistTestBlockedRepo{kind: user.KindEmail, value: "blocked@example.com"},
	}}
	_, err := app.SendRegisterCode(context.Background(), &dto.SendRegisterCodeReq{Account: "Blocked@Example.com"})
	if !errors.Is(err, errcode.ErrForbidden) {
		t.Fatalf("expected forbidden before sending code, got %v", err)
	}
}

func TestBlacklistedEmailCannotRegisterOrLogin(t *testing.T) {
	app := &UserApp{repoFactory: blacklistTestFactory{
		blockedRepo: blacklistTestBlockedRepo{kind: user.KindEmail, value: "blocked@example.com"},
	}}
	ctx := context.Background()
	if err := app.Register(ctx, &dto.RegisterReq{Account: "Blocked@Example.com"}); !errors.Is(err, errcode.ErrForbidden) {
		t.Fatalf("register: expected forbidden, got %v", err)
	}
	if _, err := app.Login(ctx, &dto.LoginReq{Account: "Blocked@Example.com", Password: "password"}); !errors.Is(err, errcode.ErrForbidden) {
		t.Fatalf("login: expected forbidden, got %v", err)
	}
}

func TestBlacklistedWechatOpenIDCannotLogin(t *testing.T) {
	oldTransport := http.DefaultClient.Transport
	http.DefaultClient.Transport = blacklistTestTransport(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"openid":"WxA"}`))}, nil
	})
	defer func() { http.DefaultClient.Transport = oldTransport }()
	app := &UserApp{
		repoFactory: blacklistTestFactory{
			blockedRepo: blacklistTestBlockedRepo{kind: user.KindOpenID, value: "WxA"},
		},
		wechatConfig: &config.WechatConfig{Jscode2SessionApi: "https://wechat.test/jscode2session"},
	}
	if _, err := app.WechatLogin(context.Background(), &dto.WechatLoginReq{Code: "test"}); !errors.Is(err, errcode.ErrForbidden) {
		t.Fatalf("wechat login: expected forbidden, got %v", err)
	}
}

func TestBlacklistedLinkedPhoneCannotRequestCode(t *testing.T) {
	email, err := user.NewEmail("blocked@example.com")
	if err != nil {
		t.Fatal(err)
	}
	app := &UserApp{repoFactory: blacklistTestFactory{
		userRepo:    blacklistTestUserRepo{current: &user.User{ID: 42, Email: email, Phone: "13800138000"}},
		blockedRepo: blacklistTestBlockedRepo{kind: user.KindEmail, value: "blocked@example.com"},
	}}
	if _, err := app.SendRegisterCode(context.Background(), &dto.SendRegisterCodeReq{Account: "13800138000"}); !errors.Is(err, errcode.ErrForbidden) {
		t.Fatalf("phone code: expected forbidden, got %v", err)
	}
}

func TestUserIDBlacklistRejectsExistingTokens(t *testing.T) {
	for _, tt := range []struct {
		name string
		user *user.User
	}{
		{name: "phone only", user: &user.User{ID: 42, Phone: "13800138000"}},
		{name: "wechat", user: &user.User{ID: 42, OpenID: "WxA"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			app := &UserApp{
				repoFactory: blacklistTestFactory{
					userRepo:    blacklistTestUserRepo{current: tt.user},
					blockedRepo: blacklistTestBlockedRepo{kind: user.KindUserID, value: "42"},
				},
				sessionRepo: blacklistTestSessionRepo{userID: tt.user.ID},
				jwtConfig:   &config.JwtConfig{JwtSecret: "blacklist-test-secret"},
			}
			access, refresh, _, err := jwt.GenerateTokenPair([]byte(app.jwtConfig.JwtSecret), time.Hour, time.Hour, tt.user.ID)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			if _, err := app.GetUserTokenClaims(ctx, access); !errors.Is(err, errcode.ErrForbidden) {
				t.Fatalf("access token: expected forbidden, got %v", err)
			}
			if _, err := app.RefreshAccessToken(ctx, refresh); !errors.Is(err, errcode.ErrForbidden) {
				t.Fatalf("refresh token: expected forbidden, got %v", err)
			}
			if blocked, err := app.IsUserBlocked(ctx, tt.user.ID); err != nil || !blocked {
				t.Fatalf("expected blocked user, got blocked %v, error %v", blocked, err)
			}
		})
	}
}

func TestUserIDBlacklistRejectsPhoneAccount(t *testing.T) {
	const phone = "13800138000"
	app := &UserApp{repoFactory: blacklistTestFactory{
		userRepo:    blacklistTestUserRepo{current: &user.User{ID: 42, Phone: phone}},
		blockedRepo: blacklistTestBlockedRepo{kind: user.KindUserID, value: "42"},
	}}
	ctx := context.Background()
	if _, err := app.SendRegisterCode(ctx, &dto.SendRegisterCodeReq{Account: phone}); !errors.Is(err, errcode.ErrForbidden) {
		t.Fatalf("phone code: expected forbidden, got %v", err)
	}
	if err := app.Register(ctx, &dto.RegisterReq{Account: phone}); !errors.Is(err, errcode.ErrForbidden) {
		t.Fatalf("register: expected forbidden, got %v", err)
	}
	if _, err := app.Login(ctx, &dto.LoginReq{Account: phone, Password: "password"}); !errors.Is(err, errcode.ErrForbidden) {
		t.Fatalf("login: expected forbidden, got %v", err)
	}
}

func TestUserIDBlacklistAllowsOtherUsers(t *testing.T) {
	app := &UserApp{
		repoFactory: blacklistTestFactory{
			userRepo:    blacklistTestUserRepo{current: &user.User{ID: 43, OpenID: "WxB"}},
			blockedRepo: blacklistTestBlockedRepo{kind: user.KindUserID, value: "42"},
		},
		sessionRepo: blacklistTestSessionRepo{userID: 43},
		jwtConfig:   &config.JwtConfig{JwtSecret: "blacklist-test-secret", JwtTimeout: time.Hour},
	}
	access, refresh, _, err := jwt.GenerateTokenPair([]byte(app.jwtConfig.JwtSecret), time.Hour, time.Hour, 43)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if claims, err := app.GetUserTokenClaims(ctx, access); err != nil || claims.UserID != 43 {
		t.Fatalf("expected valid access token claims, got %v, error %v", claims, err)
	}
	if tokens, err := app.RefreshAccessToken(ctx, refresh); err != nil || tokens.AccessToken == "" {
		t.Fatalf("expected refreshed access token, got %v, error %v", tokens, err)
	}
	if blocked, err := app.IsUserBlocked(ctx, 43); err != nil || blocked {
		t.Fatalf("expected allowed user, got blocked %v, error %v", blocked, err)
	}
}
