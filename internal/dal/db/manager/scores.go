package manager

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/dal/db/query"
	"github.com/superwhys/billiard-helper/internal/models/dbmodels"
	"github.com/superwhys/billiard-helper/internal/ports"
	"gorm.io/gorm"
)

type scoresManager struct {
	query *query.Query
}

func NewScoresManager(db *gorm.DB) ports.ScoreRepo {
	return &scoresManager{query: query.Use(db)}
}

func (m *scoresManager) CreateScore(ctx context.Context, score *dbmodels.Scores) error {
	return m.query.Scores.WithContext(ctx).Create(score)
}

func (m *scoresManager) DeleteScore(ctx context.Context, scoreID uint) error {
	s := m.query.Scores
	_, err := s.WithContext(ctx).
		Where(s.ID.Eq(scoreID)).
		Delete()
	return err
}
