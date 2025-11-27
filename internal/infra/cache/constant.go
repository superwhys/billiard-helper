package cache

import "time"

var (
	EmailCodeLock          = genCacheWithKey("email_code:lock", 30*time.Second)
	EmailCodeCache         = genCacheWithKey("email_code", 10*time.Minute)
	EmailCodeCooldownCache = genCacheWithKey("email_code:cooldown", time.Minute)

	RoomLockCache = genCacheWithKey("room:lock", 5*time.Second)
)
