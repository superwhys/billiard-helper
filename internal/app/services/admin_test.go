package services

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/app/factory"
	"github.com/superwhys/billiard-helper/internal/constant"
	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/domain/shared"
	"github.com/superwhys/billiard-helper/internal/domain/user"
	"github.com/superwhys/billiard-helper/internal/errcode"
)

type adminTestFactory struct {
	factory.IRepoFactory
	users   *adminTestUsers
	blocked *adminTestBlocked
	matches *adminTestMatches
}

func (f adminTestFactory) UserRepo() user.IUserRepository                       { return f.users }
func (f adminTestFactory) BlockedIdentityRepo() user.IBlockedIdentityRepository { return f.blocked }
func (f adminTestFactory) MatchRepo() match.IMatchRepository                    { return f.matches }

type adminTestUsers struct {
	user.IUserRepository
	items         []*user.User
	offset, limit int
}

func (r *adminTestUsers) List(_ context.Context, offset, limit int) ([]*user.User, int64, error) {
	r.offset, r.limit = offset, limit
	return r.items, int64(len(r.items)), nil
}

func (r *adminTestUsers) FindByID(_ context.Context, id uint) (*user.User, error) {
	return &user.User{ID: id, Phone: "13800138000"}, nil
}

type adminTestBlocked struct {
	user.IBlockedIdentityRepository
	kind  user.BlockedIdentityKind
	value string
	err   error
}

func (r *adminTestBlocked) Block(_ context.Context, kind user.BlockedIdentityKind, value string) error {
	if r.err != nil {
		return r.err
	}
	r.kind, r.value = kind, value
	return nil
}

func (r *adminTestBlocked) IsBlocked(_ context.Context, kind user.BlockedIdentityKind, value string) (bool, error) {
	return r.kind == kind && r.value == value, r.err
}

type adminTestBus struct {
	shared.EventBus
	message *shared.QueueMessage
	channel string
	err     error
}

func (b *adminTestBus) Publish(_ context.Context, channel string, msg *shared.QueueMessage) error {
	b.channel, b.message = channel, msg
	return b.err
}

type adminTestMatches struct {
	match.IMatchRepository
	start, end time.Time
}

func (r *adminTestMatches) CountCreatedBetween(_ context.Context, start, end time.Time) (*match.CreationStats, error) {
	r.start, r.end = start, end
	return &match.CreationStats{Total: 3, ByType: map[match.MatchType]int64{match.MatchTypeSnooker: 3}}, nil
}

func TestSuperAdminOnlyComesFromConfiguredID(t *testing.T) {
	for _, tt := range []struct {
		configured, current uint
		want                bool
	}{
		{0, 0, false}, {0, 7, false}, {7, 0, false}, {7, 8, false}, {7, 7, true},
	} {
		app := NewAdminApp(nil, nil, nil, tt.configured)
		if got := app.IsSuperAdmin(tt.current); got != tt.want {
			t.Errorf("configured=%d current=%d: got %v", tt.configured, tt.current, got)
		}
	}
}

func TestAdminOperationsRejectNonAdmin(t *testing.T) {
	app := NewAdminApp(nil, nil, nil, 7)
	ctx := context.Background()
	if _, err := app.ListUsers(ctx, 8, &dto.AdminUsersRequest{}); !errors.Is(err, errcode.ErrForbidden) {
		t.Fatalf("list: %v", err)
	}
	if _, err := app.BlockUser(ctx, 8, 42); !errors.Is(err, errcode.ErrForbidden) {
		t.Fatalf("block: %v", err)
	}
	if _, err := app.TodayMatchStats(ctx, 8); !errors.Is(err, errcode.ErrForbidden) {
		t.Fatalf("stats: %v", err)
	}
}

func TestAdminCannotBlockSelfOrZeroID(t *testing.T) {
	app := NewAdminApp(nil, nil, nil, 7)
	for _, id := range []uint{0, 7} {
		if _, err := app.BlockUser(context.Background(), 7, id); !errors.Is(err, errcode.ErrBadRequest) {
			t.Fatalf("target %d: %v", id, err)
		}
	}
}

func TestAdminBlockPhoneOnlyUserPersistsBeforeNotification(t *testing.T) {
	for _, publishErr := range []error{nil, errors.New("redis unavailable")} {
		repos := adminTestFactory{users: &adminTestUsers{}, blocked: &adminTestBlocked{}}
		bus := &adminTestBus{err: publishErr}
		app := NewAdminApp(repos, nil, bus, 7)
		result, err := app.BlockUser(context.Background(), 7, 42)
		if err != nil {
			t.Fatal(err)
		}
		if !result.IsBlocked || result.DisconnectNotified != (publishErr == nil) {
			t.Fatalf("unexpected result: %+v", result)
		}
		if repos.blocked.kind != user.KindUserID || repos.blocked.value != "42" {
			t.Fatal("user ID was not blocked")
		}
		if bus.channel != constant.BilliardEventChannel || bus.message.Event != constant.EventUserBlocked {
			t.Fatal("wrong disconnect event")
		}
		var payload struct {
			UserID uint `json:"user_id"`
		}
		if err := json.Unmarshal(bus.message.Data, &payload); err != nil || payload.UserID != 42 {
			t.Fatalf("bad payload: %s", bus.message.Data)
		}
	}
}

func TestAdminPaginationAndBlacklistStatus(t *testing.T) {
	repos := adminTestFactory{
		users:   &adminTestUsers{items: []*user.User{{ID: 42, Name: "blocked"}, {ID: 7, Name: "admin"}}},
		blocked: &adminTestBlocked{kind: user.KindUserID, value: "42"},
	}
	app := NewAdminApp(repos, &UserApp{repoFactory: repos}, nil, 7)
	result, err := app.ListUsers(context.Background(), 7, &dto.AdminUsersRequest{Page: 2, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if repos.users.offset != 10 || repos.users.limit != 10 {
		t.Fatal("wrong page range")
	}
	if !result.Items[0].IsBlocked || result.Items[0].IsSuperAdmin || !result.Items[1].IsSuperAdmin {
		t.Fatalf("wrong statuses: %+v", result.Items)
	}
	for _, req := range []dto.AdminUsersRequest{{Page: -1}, {PageSize: -1}, {PageSize: 101}, {Page: int(^uint(0) >> 1), PageSize: 100}} {
		if _, err := app.ListUsers(context.Background(), 7, &req); !errors.Is(err, errcode.ErrBadRequest) {
			t.Fatalf("invalid pagination %+v: %v", req, err)
		}
	}
}

func TestBeijingDayBounds(t *testing.T) {
	for _, tt := range []struct{ input, wantDate string }{
		{"2026-09-28T15:59:59Z", "2026-09-28"},
		{"2026-09-28T16:00:00Z", "2026-09-29"},
	} {
		now, _ := time.Parse(time.RFC3339, tt.input)
		start, end := beijingDayBounds(now)
		if start.Format("2006-01-02") != tt.wantDate || start.Hour() != 0 || end.Sub(start) != 24*time.Hour {
			t.Fatalf("unexpected range %v / %v", start, end)
		}
		_, offset := start.Zone()
		if offset != 8*60*60 {
			t.Fatal("expected UTC+8 day boundary")
		}
	}
}

func TestAdminStatsKeepZeroTypes(t *testing.T) {
	repos := adminTestFactory{matches: &adminTestMatches{}}
	result, err := NewAdminApp(repos, nil, nil, 7).TodayMatchStats(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 3 || result.ByType["snooker"] != 3 || len(result.ByType) != 3 || result.Timezone != "Asia/Shanghai" {
		t.Fatalf("bad stats: %+v", result)
	}
}
