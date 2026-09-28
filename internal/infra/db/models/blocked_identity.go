package models

import "time"

type BlockedIdentity struct {
	ID        uint      `gorm:"primaryKey"`
	Kind      string    `gorm:"column:kind;type:VARCHAR(16);not null;uniqueIndex:idx_blocked_identity,priority:1"`
	Value     string    `gorm:"column:value;type:VARBINARY(255);not null;uniqueIndex:idx_blocked_identity,priority:2"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (b *BlockedIdentity) TableName() string {
	return "blocked_identities"
}
