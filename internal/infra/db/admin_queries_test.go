package db

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/infra/db/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type adminQueryStatement struct {
	sql  string
	vars []interface{}
}

func newAdminQueryDryRunDB(t *testing.T) (*gorm.DB, *[]adminQueryStatement) {
	t.Helper()
	database, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "unused:unused@tcp(127.0.0.1:1)/unused",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true, Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	var statements []adminQueryStatement
	capture := func(tx *gorm.DB) {
		statements = append(statements, adminQueryStatement{
			sql:  tx.Statement.SQL.String(),
			vars: append([]interface{}(nil), tx.Statement.Vars...),
		})
	}
	if err := database.Callback().Query().After("gorm:query").Register("test:capture_admin_query", capture); err != nil {
		t.Fatal(err)
	}
	if err := database.Callback().Row().After("gorm:row").Register("test:capture_admin_row", capture); err != nil {
		t.Fatal(err)
	}
	return database, &statements
}

func TestUserListSelectsPublicFieldsAndPaginates(t *testing.T) {
	database, statements := newAdminQueryDryRunDB(t)
	if _, _, err := NewUserRepo(database).List(context.Background(), 20, 10); err != nil {
		t.Fatal(err)
	}
	if len(*statements) != 2 {
		t.Fatalf("expected count and page queries, got %d", len(*statements))
	}
	count, page := (*statements)[0], (*statements)[1]
	if count.sql != "SELECT count(*) FROM `users` WHERE `users`.`deleted_at` IS NULL" {
		t.Fatalf("unexpected count query: %s", count.sql)
	}
	for _, fragment := range []string{"`users`.`email`", "`users`.`open_id`", "`users`.`created_at`", "ORDER BY `users`.`id` DESC LIMIT ? OFFSET ?"} {
		if !strings.Contains(page.sql, fragment) {
			t.Fatalf("page query missing %q: %s", fragment, page.sql)
		}
	}
	if strings.Contains(page.sql, "password") || strings.Contains(page.sql, "SELECT *") {
		t.Fatalf("page query must exclude password: %s", page.sql)
	}
	if !reflect.DeepEqual(page.vars, []interface{}{10, 20}) {
		t.Fatalf("unexpected pagination parameters: %v", page.vars)
	}
}

func TestMatchCreationStatsQueryUsesHalfOpenRangeAndIncludesDeleted(t *testing.T) {
	database, statements := newAdminQueryDryRunDB(t)
	start := time.Date(2026, 9, 28, 0, 0, 0, 0, time.FixedZone("Asia/Shanghai", 8*60*60))
	end := start.AddDate(0, 0, 1)
	_, err := NewMatchRepo(database).CountCreatedBetween(context.Background(), start, end)
	if !errors.Is(err, gorm.ErrDryRunModeUnsupported) {
		t.Fatalf("Scan must stop without opening a real connection, got %v", err)
	}
	if len(*statements) != 1 {
		t.Fatalf("expected one aggregate query, got %d", len(*statements))
	}
	statement := (*statements)[0]
	want := "SELECT `matches`.`match_type`,COUNT(`matches`.`id`) AS `count` FROM `matches` WHERE `matches`.`created_at` >= ? AND `matches`.`created_at` < ? GROUP BY `matches`.`match_type`"
	if statement.sql != want {
		t.Fatalf("unexpected stats query:\n got: %s\nwant: %s", statement.sql, want)
	}
	if !reflect.DeepEqual(statement.vars, []interface{}{start, end}) {
		t.Fatalf("unexpected range parameters: %v", statement.vars)
	}
}

func TestAdminQueriesMySQL(t *testing.T) {
	dsn := os.Getenv("BILLIARD_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set BILLIARD_TEST_MYSQL_DSN for an isolated test database")
	}
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := database.AutoMigrate(models.Tables()...); err != nil {
		t.Fatal(err)
	}
	tx := database.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	t.Cleanup(func() { tx.Rollback() })
	ctx := context.Background()
	userRepo := NewUserRepo(tx)
	_, initialUserCount, err := userRepo.List(ctx, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	createdAt := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)
	users := []*models.User{
		{Model: gorm.Model{CreatedAt: createdAt}, Name: "first", Password: "must-not-be-loaded"},
		{Name: "second"},
		{Name: "deleted"},
	}
	if err := tx.Create(users).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Delete(users[2]).Error; err != nil {
		t.Fatal(err)
	}
	page, total, err := userRepo.List(ctx, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if total != initialUserCount+2 || len(page) != 1 || page[0].ID != users[0].ID {
		t.Fatalf("unexpected page or total: page=%+v total=%d", page, total)
	}
	if page[0].Password.Hash() != "" || !page[0].CreatedAt.Equal(createdAt) {
		t.Fatalf("password must be omitted and creation time preserved: %+v", page[0])
	}
	page, total, err = userRepo.List(ctx, int(total), 1)
	if err != nil || len(page) != 0 || total != initialUserCount+2 {
		t.Fatalf("out-of-range page: count=%d total=%d err=%v", len(page), total, err)
	}

	matchRepo := NewMatchRepo(tx)
	start := time.Date(2098, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	initialStats, err := matchRepo.CountCreatedBetween(ctx, start, end)
	if err != nil {
		t.Fatal(err)
	}
	matches := []*models.Match{
		{Model: gorm.Model{CreatedAt: start.Add(-time.Second)}, UserID: users[0].ID, Name: "before", MatchType: "snooker"},
		{Model: gorm.Model{CreatedAt: start}, UserID: users[0].ID, Name: "start", MatchType: "snooker"},
		{Model: gorm.Model{CreatedAt: end.Add(-time.Second)}, UserID: users[0].ID, Name: "last", MatchType: "snooker"},
		{Model: gorm.Model{CreatedAt: start.Add(time.Hour)}, UserID: users[0].ID, Name: "deleted", MatchType: "other"},
		{Model: gorm.Model{CreatedAt: end}, UserID: users[0].ID, Name: "end", MatchType: "other"},
	}
	if err := tx.Create(matches).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Delete(matches[3]).Error; err != nil {
		t.Fatal(err)
	}
	stats, err := matchRepo.CountCreatedBetween(ctx, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Total != initialStats.Total+3 || stats.ByType[match.MatchTypeSnooker] != initialStats.ByType[match.MatchTypeSnooker]+2 || stats.ByType["other"] != initialStats.ByType["other"]+1 {
		t.Fatalf("unexpected creation counts: %+v (before: %+v)", stats, initialStats)
	}
	empty, err := matchRepo.CountCreatedBetween(ctx, start, start)
	if err != nil || empty.Total != 0 || len(empty.ByType) != 0 {
		t.Fatalf("empty range: stats=%+v err=%v", empty, err)
	}
}
