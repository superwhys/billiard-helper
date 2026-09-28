package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/superwhys/billiard-helper/api"
	"github.com/superwhys/billiard-helper/config"
	"github.com/superwhys/billiard-helper/internal/app/factory"
	"github.com/superwhys/billiard-helper/internal/app/hook"
	"github.com/superwhys/billiard-helper/internal/app/services"
	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/domain/shared"
	"github.com/superwhys/billiard-helper/internal/domain/user"
	"github.com/superwhys/billiard-helper/internal/errcode"
	"github.com/superwhys/billiard-helper/internal/infra/socket"
	"github.com/superwhys/billiard-helper/internal/pkg/jwt"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"gorm.io/gorm"
)

type adminHTTPRepos struct {
	factory.IRepoFactory
	users   *adminHTTPUsers
	blocked *adminHTTPBlocked
	matches adminHTTPMatches
}

func (r adminHTTPRepos) UserRepo() user.IUserRepository                       { return r.users }
func (r adminHTTPRepos) BlockedIdentityRepo() user.IBlockedIdentityRepository { return r.blocked }
func (r adminHTTPRepos) MatchRepo() match.IMatchRepository                    { return r.matches }

type adminHTTPUsers struct {
	user.IUserRepository
	users map[uint]*user.User
}

func (r *adminHTTPUsers) FindByID(_ context.Context, id uint) (*user.User, error) {
	u, exists := r.users[id]
	if !exists {
		return nil, gorm.ErrRecordNotFound
	}
	return u, nil
}

func (r *adminHTTPUsers) List(context.Context, int, int) ([]*user.User, int64, error) {
	return []*user.User{r.users[2], r.users[1]}, 2, nil
}

type adminHTTPBlocked struct {
	user.IBlockedIdentityRepository
	values map[string]bool
}

func (r *adminHTTPBlocked) IsBlocked(_ context.Context, kind user.BlockedIdentityKind, value string) (bool, error) {
	return r.values[string(kind)+":"+value], nil
}

func (r *adminHTTPBlocked) Block(_ context.Context, kind user.BlockedIdentityKind, value string) error {
	r.values[string(kind)+":"+value] = true
	return nil
}

type adminHTTPMatches struct{ match.IMatchRepository }

func (adminHTTPMatches) CountCreatedBetween(context.Context, time.Time, time.Time) (*match.CreationStats, error) {
	return &match.CreationStats{Total: 1, ByType: map[match.MatchType]int64{match.MatchType8Ball: 1}}, nil
}

type adminHTTPSessions struct {
	user.ISessionRepository
	values map[string]uint
}

func (s adminHTTPSessions) GetSession(_ context.Context, id string) (uint, error) {
	if userID, exists := s.values[id]; exists {
		return userID, nil
	}
	return 0, errcode.ErrTokenExpired
}

type adminHTTPBus struct {
	shared.EventBus
	events int
}

func (b *adminHTTPBus) Publish(context.Context, string, *shared.QueueMessage) error {
	b.events++
	return nil
}

func TestAdminHTTPAccessAndBlocking(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repos := adminHTTPRepos{
		users: &adminHTTPUsers{users: map[uint]*user.User{
			1: {ID: 1, Name: "admin", Password: user.NewPasswordFromHash("secret-hash")},
			2: {ID: 2, Name: "member", Phone: "13800138000"},
		}},
		blocked: &adminHTTPBlocked{values: make(map[string]bool)},
	}
	sessions := adminHTTPSessions{values: make(map[string]uint)}
	jwtConf := &config.JwtConfig{JwtSecret: "admin-http-test", JwtTimeout: time.Hour, JwtRefreshTimeout: 24 * time.Hour}
	tokens := make(map[uint]string)
	for _, id := range []uint{1, 2} {
		access, _, session, err := jwt.GenerateTokenPair([]byte(jwtConf.JwtSecret), time.Hour, 24*time.Hour, id)
		if err != nil {
			t.Fatal(err)
		}
		tokens[id], sessions.values[session] = access, id
	}
	userApp := services.NewUserApp(nil, repos, sessions, nil, nil, nil, jwtConf, nil)
	bus := &adminHTTPBus{}
	adminApp := services.NewAdminApp(repos, userApp, bus, 1)
	socketManager := socket.NewSocketManager(hook.NewSocketHook(nil, userApp))
	handler := api.SetupApi(false, socketManager, limiter.New(memory.NewStore(), limiter.Rate{Period: time.Minute, Limit: 1000}), userApp, nil, nil, nil, adminApp)
	call := func(method, path, token, body string) (*httptest.ResponseRecorder, struct {
		Code int
		Data json.RawMessage
	}) { t.Helper(); req := httptest.NewRequest(method, path, strings.NewReader(body)); req.Header.Set("Content-Type", "application/json"); if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}; rec := httptest.NewRecorder(); handler.ServeHTTP(rec, req); var payload struct {
		Code int
		Data json.RawMessage
	}; if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON: %s", rec.Body.String())
	}; return rec, payload }
	for _, tt := range []struct {
		token string
		want  bool
	}{{"", false}, {tokens[2], false}, {tokens[1], true}} {
		rec, payload := call(http.MethodGet, "/app/state?user_id=1&is_super_admin=true", tt.token, "")
		var state struct {
			IsSuperAdmin bool `json:"is_super_admin"`
		}
		if err := json.Unmarshal(payload.Data, &state); err != nil {
			t.Fatal(err)
		}
		if rec.Code != 200 || payload.Code != 0 || state.IsSuperAdmin != tt.want {
			t.Fatalf("unexpected state: %s", rec.Body.String())
		}
		if bytes.Contains(payload.Data, []byte("super_admin_user_id")) {
			t.Fatal("state leaked configured admin ID")
		}
	}
	if rec, _ := call(http.MethodGet, "/app/state", "forged-token", ""); rec.Code != http.StatusUnauthorized {
		t.Fatal("invalid token accepted by state endpoint")
	}
	for _, endpoint := range []struct{ method, path, body string }{
		{http.MethodGet, "/admin/users", ""},
		{http.MethodGet, "/admin/matches/stats", ""},
		{http.MethodPost, "/admin/users/block", `{"user_id":2,"is_super_admin":true}`},
	} {
		for _, tt := range []struct {
			token  string
			status int
		}{{"", http.StatusUnauthorized}, {tokens[2], http.StatusForbidden}} {
			if rec, _ := call(endpoint.method, endpoint.path, tt.token, endpoint.body); rec.Code != tt.status {
				t.Fatalf("permission failure %s: %d", endpoint.path, rec.Code)
			}
		}
	}
	if len(repos.blocked.values) != 0 {
		t.Fatal("non-admin changed blacklist")
	}
	rec, payload := call(http.MethodGet, "/admin/users", tokens[1], "")
	if payload.Code != 0 || bytes.Contains(rec.Body.Bytes(), []byte("secret-hash")) || bytes.Contains(rec.Body.Bytes(), []byte("password")) {
		t.Fatalf("unsafe user list: %s", rec.Body.String())
	}
	if _, payload = call(http.MethodGet, "/admin/matches/stats", tokens[1], ""); payload.Code != 0 {
		t.Fatal("admin cannot read statistics")
	}
	if _, payload = call(http.MethodPost, "/admin/users/block", tokens[1], `{"user_id":1}`); payload.Code != errcode.ErrBadRequest.Code() {
		t.Fatal("admin was allowed to block itself")
	}
	if _, payload = call(http.MethodPost, "/admin/users/block", tokens[1], `{"user_id":999}`); payload.Code != errcode.ErrCodeUserNotFound.Code() {
		t.Fatal("missing user was accepted")
	}
	if _, payload = call(http.MethodPost, "/admin/users/block", tokens[1], `{"user_id":2}`); payload.Code != 0 || bus.events != 1 {
		t.Fatal("admin block failed")
	}
	if rec, payload = call(http.MethodGet, "/app/state", tokens[2], ""); rec.Code != http.StatusForbidden || payload.Code != errcode.ErrForbidden.Code() {
		t.Fatal("blocked account kept access with existing token")
	}
	if rec, _ = call(http.MethodGet, "/match/list", tokens[2], ""); rec.Code != http.StatusForbidden {
		t.Fatal("blocked account can access business routes")
	}
}
