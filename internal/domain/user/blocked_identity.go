package user

import (
	"fmt"
	"strconv"
	"strings"
)

type BlockedIdentityKind string

const (
	KindEmail  BlockedIdentityKind = "email"
	KindOpenID BlockedIdentityKind = "open_id"
	KindUserID BlockedIdentityKind = "user_id"
)

func NormalizeBlockedIdentity(kind BlockedIdentityKind, value string) (string, error) {
	if value == "" || len(value) > 255 || strings.TrimSpace(value) != value {
		return "", fmt.Errorf("invalid blocked identity value")
	}

	switch kind {
	case KindEmail:
		if _, err := NewEmail(value); err != nil {
			return "", err
		}
		return strings.ToLower(value), nil
	case KindOpenID:
		return value, nil
	case KindUserID:
		id, err := strconv.ParseUint(value, 10, strconv.IntSize)
		if err != nil || id == 0 {
			return "", fmt.Errorf("invalid blocked user id")
		}
		return strconv.FormatUint(id, 10), nil
	default:
		return "", fmt.Errorf("invalid blocked identity kind: %s", kind)
	}
}
