package codegen

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/superwhys/billiard-helper/internal/models/types"
)

// GenerateHash 根据输入内容生成唯一且幂等的值
func GenerateHash(contents ...string) string {
	hasher := sha256.New()
	for _, content := range contents {
		hasher.Write([]byte(content))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func GeneratePlayerCode(roomID uint, playerType types.PlayerType, playerNickName string) string {
	code := GenerateHash(
		fmt.Sprintf("%d", roomID),
		fmt.Sprintf("%d", playerType),
		fmt.Sprintf("%s", playerNickName),
	)
	return code
}
