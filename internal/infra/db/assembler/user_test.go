package assembler

import (
	"testing"
	"time"

	"github.com/superwhys/billiard-helper/internal/infra/db/models"
	"gorm.io/gorm"
)

func TestUserCreationTimeRoundTrip(t *testing.T) {
	createdAt := time.Date(2026, 9, 28, 1, 2, 3, 0, time.UTC)
	po := &models.User{Model: gorm.Model{ID: 42, CreatedAt: createdAt}, Name: "user"}
	assembler := NewUserPoAssembler()
	entity := assembler.ToEntity(po)
	if !entity.CreatedAt.Equal(createdAt) {
		t.Fatalf("entity creation time: got %v, want %v", entity.CreatedAt, createdAt)
	}
	if result := assembler.ToPO(entity); !result.CreatedAt.Equal(createdAt) {
		t.Fatalf("PO creation time: got %v, want %v", result.CreatedAt, createdAt)
	}
}
