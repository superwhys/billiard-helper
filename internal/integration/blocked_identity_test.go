package integration_test

import (
	"context"
	"os"
	"testing"

	"github.com/superwhys/billiard-helper/internal/domain/user"
	"github.com/superwhys/billiard-helper/internal/infra/db"
	"github.com/superwhys/billiard-helper/internal/infra/db/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestBlockedIdentityRepository(t *testing.T) {
	dsn := os.Getenv("BILLIARD_TEST_MYSQL_DSN")
	if dsn == "" {
		t.Skip("set BILLIARD_TEST_MYSQL_DSN for an isolated test database")
	}

	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&models.BlockedIdentity{}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	repo := db.NewBlockedIdentityRepo(database)
	email, openID := "Blocked.Test@Example.com", "WxAbC123"
	t.Cleanup(func() {
		_ = repo.Unblock(ctx, user.KindEmail, email)
		_ = repo.Unblock(ctx, user.KindOpenID, openID)
	})

	if err := repo.Block(ctx, user.KindEmail, email); err != nil {
		t.Fatal(err)
	}
	if err := repo.Block(ctx, user.KindEmail, "blocked.test@example.com"); err != nil {
		t.Fatalf("duplicate block should be idempotent: %v", err)
	}
	var count int64
	if err := database.Model(&models.BlockedIdentity{}).Where("kind = ? AND value = ?", string(user.KindEmail), "blocked.test@example.com").Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("duplicate email block created %d rows", count)
	}
	if blocked, err := repo.IsBlocked(ctx, user.KindEmail, "BLOCKED.TEST@EXAMPLE.COM"); err != nil || !blocked {
		t.Fatalf("email case normalization failed: blocked=%v err=%v", blocked, err)
	}
	if err := repo.Block(ctx, user.KindOpenID, openID); err != nil {
		t.Fatal(err)
	}
	if blocked, err := repo.IsBlocked(ctx, user.KindOpenID, "wxAbC123"); err != nil || blocked {
		t.Fatalf("open_id comparison must be case-sensitive: blocked=%v err=%v", blocked, err)
	}
	if err := repo.Unblock(ctx, user.KindEmail, email); err != nil {
		t.Fatal(err)
	}
	if err := repo.Unblock(ctx, user.KindEmail, email); err != nil {
		t.Fatalf("duplicate unblock should be idempotent: %v", err)
	}
	if blocked, err := repo.IsBlocked(ctx, user.KindEmail, email); err != nil || blocked {
		t.Fatalf("unblock failed: blocked=%v err=%v", blocked, err)
	}
	if err := repo.Block(ctx, "phone", "123456789"); err == nil {
		t.Fatal("invalid kind must be rejected")
	}
}
