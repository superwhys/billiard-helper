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

func (m *scoresManager) GetLatestScore(ctx context.Context, roomID, playerID uint) (*dbmodels.Scores, error) {
	s := m.query.Scores
	return s.WithContext(ctx).
		Where(s.RoomID.Eq(roomID), s.PlayerID.Eq(playerID)).
		Order(s.ID.Desc()).
		First()
}

func (m *scoresManager) GetRoomScores(ctx context.Context, roomID uint) ([]*dbmodels.Scores, error) {
	s := m.query.Scores
	return s.WithContext(ctx).
		Where(s.RoomID.Eq(roomID)).
		Order(s.ID.Asc()). // 按照发生顺序返回
		Preload(s.Operator).
		Find()
}
