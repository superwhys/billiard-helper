package user

import (
	"fmt"
	"strings"
)

type BlockedIdentityKind string

const (
	KindEmail  BlockedIdentityKind = "email"
	KindOpenID BlockedIdentityKind = "open_id"
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
	default:
		return "", fmt.Errorf("invalid blocked identity kind: %s", kind)
	}
}
