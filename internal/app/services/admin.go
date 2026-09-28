package services

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/app/factory"
	"github.com/superwhys/billiard-helper/internal/constant"
	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/domain/shared"
	"github.com/superwhys/billiard-helper/internal/domain/user"
	"github.com/superwhys/billiard-helper/internal/errcode"
	"gorm.io/gorm"
)

type AdminApp struct {
	repoFactory      factory.IRepoFactory
	userApp          *UserApp
	eventBus         shared.EventBus
	superAdminUserID uint
}

func NewAdminApp(repoFactory factory.IRepoFactory, userApp *UserApp, eventBus shared.EventBus, superAdminUserID uint) *AdminApp {
	return &AdminApp{
		repoFactory:      repoFactory,
		userApp:          userApp,
		eventBus:         eventBus,
		superAdminUserID: superAdminUserID,
	}
}

func (a *AdminApp) IsSuperAdmin(userID uint) bool {
	return a.superAdminUserID != 0 && userID == a.superAdminUserID
}

func (a *AdminApp) ListUsers(ctx context.Context, operatorID uint, req *dto.AdminUsersRequest) (*dto.AdminUsersResponse, error) {
	if !a.IsSuperAdmin(operatorID) {
		return nil, errcode.ErrForbidden
	}
	page, pageSize := req.Page, req.PageSize
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}
	if page < 1 || pageSize < 1 || pageSize > 100 || page-1 > int(^uint(0)>>1)/pageSize {
		return nil, errcode.ErrBadRequest
	}
	users, total, err := a.repoFactory.UserRepo().List(ctx, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]dto.AdminUser, 0, len(users))
	for _, u := range users {
		err := a.userApp.checkBlockedUserIdentity(ctx, u)
		blocked := errors.Is(err, errcode.ErrForbidden)
		if err != nil && !blocked {
			return nil, err
		}
		items = append(items, dto.AdminUser{
			ID:           u.ID,
			Name:         u.Name,
			Email:        u.Email.String(),
			OpenID:       u.OpenID,
			Phone:        u.Phone,
			Avatar:       u.Avatar,
			CreatedAt:    u.CreatedAt,
			IsBlocked:    blocked,
			IsSuperAdmin: a.IsSuperAdmin(u.ID),
		})
	}
	return &dto.AdminUsersResponse{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

func (a *AdminApp) BlockUser(ctx context.Context, operatorID, userID uint) (*dto.BlockUserResponse, error) {
	if !a.IsSuperAdmin(operatorID) {
		return nil, errcode.ErrForbidden
	}
	if userID == 0 {
		return nil, errcode.ErrBadRequest
	}
	if a.IsSuperAdmin(userID) {
		return nil, errcode.ErrBadRequest.WithMessage("不能将超管加入黑名单")
	}
	if _, err := a.repoFactory.UserRepo().FindByID(ctx, userID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrCodeUserNotFound
		}
		return nil, err
	}
	if err := a.repoFactory.BlockedIdentityRepo().Block(ctx, user.KindUserID, strconv.FormatUint(uint64(userID), 10)); err != nil {
		return nil, err
	}

	payload, err := json.Marshal(struct {
		UserID uint `json:"user_id"`
	}{UserID: userID})
	if err != nil {
		return nil, err
	}
	err = a.eventBus.Publish(ctx, constant.BilliardEventChannel, &shared.QueueMessage{Event: constant.EventUserBlocked, Data: payload})
	if err != nil {
		logging.Errorc(ctx, "user(%d) blocked by admin(%d), disconnect notification failed: %v", userID, operatorID, err)
	}
	return &dto.BlockUserResponse{IsBlocked: true, DisconnectNotified: err == nil}, nil
}

func (a *AdminApp) TodayMatchStats(ctx context.Context, operatorID uint) (*dto.AdminMatchStats, error) {
	if !a.IsSuperAdmin(operatorID) {
		return nil, errcode.ErrForbidden
	}
	start, end := beijingDayBounds(time.Now())
	stats, err := a.repoFactory.MatchRepo().CountCreatedBetween(ctx, start, end)
	if err != nil {
		return nil, err
	}
	byType := map[string]int64{
		string(match.MatchType8Ball):   0,
		string(match.MatchType9Ball):   0,
		string(match.MatchTypeSnooker): 0,
	}
	for kind, count := range stats.ByType {
		byType[string(kind)] = count
	}
	return &dto.AdminMatchStats{
		Date:     start.Format("2006-01-02"),
		Timezone: "Asia/Shanghai",
		Total:    stats.Total,
		ByType:   byType,
	}, nil
}

func beijingDayBounds(now time.Time) (time.Time, time.Time) {
	local := now.In(time.FixedZone("Asia/Shanghai", 8*60*60))
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location())
	return start, start.AddDate(0, 0, 1)
}
