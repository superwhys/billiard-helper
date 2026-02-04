package cache

import (
	"time"
)

var (
	VerifyCodeLock          = genCacheWithKey("verify_code:lock", 30*time.Second)
	VerifyCodeCache         = genCacheWithKey("verify_code", 10*time.Minute)
	VerifyCodeCooldownCache = genCacheWithKey("verify_code:cooldown", time.Minute)

	MatchLockCache = genCacheWithKey("match:lock", 5*time.Second)

	// AuthSessionCache stores the currently valid token for a user (Single Device Login)
	AuthSessionCache = genCacheWithKey("auth:session", 24*time.Hour)
)
