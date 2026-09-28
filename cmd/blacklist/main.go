package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/miebyte/goutils/flags"
	"github.com/miebyte/goutils/mysqlutils"
	"github.com/miebyte/goutils/redisutils"
	"github.com/superwhys/billiard-helper/internal/constant"
	"github.com/superwhys/billiard-helper/internal/domain/shared"
	"github.com/superwhys/billiard-helper/internal/domain/user"
	"github.com/superwhys/billiard-helper/internal/infra/db"
	"github.com/superwhys/billiard-helper/internal/infra/eventbus"
	"gorm.io/gorm"
)

var (
	actionFlag      = flags.String("action", "", "block or unblock")
	kindFlag        = flags.String("kind", "", "email or open_id")
	valueFlag       = flags.String("value", "", "email address or WeChat open_id")
	mysqlConfigFlag = flags.Struct("mysql", (*mysqlutils.MysqlConfig)(nil), "mysql config")
	redisConfigFlag = flags.Struct("redis", (*redisutils.RedisConfig)(nil), "redis config")
)

func main() {
	flags.Parse()
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	action := actionFlag()
	if action != "block" && action != "unblock" {
		return fmt.Errorf("action must be block or unblock")
	}
	kind := user.BlockedIdentityKind(kindFlag())
	value, err := user.NormalizeBlockedIdentity(kind, valueFlag())
	if err != nil {
		return err
	}

	mysqlConf := new(mysqlutils.MysqlConfig)
	if err := mysqlConfigFlag(mysqlConf); err != nil {
		return err
	}
	mysqlDB, err := mysqlConf.DialMysqlGorm()
	if err != nil {
		return err
	}

	blockedRepo := db.NewBlockedIdentityRepo(mysqlDB)
	if action == "unblock" {
		if err := blockedRepo.Unblock(ctx, kind, value); err != nil {
			return err
		}
		log.Printf("unblocked %s: %s", kind, value)
		return nil
	}

	redisConf := new(redisutils.RedisConfig)
	if err := redisConfigFlag(redisConf); err != nil {
		return err
	}
	redisClient, err := redisConf.DialGORedisClient()
	if err != nil {
		return err
	}
	defer redisClient.Close()

	userRepo := db.NewUserRepo(mysqlDB)
	var existingUser *user.User
	switch kind {
	case user.KindEmail:
		existingUser, err = userRepo.FindByEmailFold(ctx, value)
	case user.KindOpenID:
		existingUser, err = userRepo.FindByOpenID(ctx, value)
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err := blockedRepo.Block(ctx, kind, value); err != nil {
		return err
	}

	if existingUser != nil {
		payload, err := json.Marshal(struct {
			UserID uint `json:"user_id"`
		}{UserID: existingUser.ID})
		if err != nil {
			return err
		}
		bus := eventbus.NewRedisEventBus(redisClient)
		if err := bus.Publish(ctx, constant.BilliardEventChannel, &shared.QueueMessage{
			Event: constant.EventUserBlocked,
			Data:  payload,
		}); err != nil {
			return fmt.Errorf("blacklist saved, websocket disconnect notification failed: %w", err)
		}
	}
	log.Printf("blocked %s: %s", kind, value)
	return nil
}
