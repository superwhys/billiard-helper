package manager

import (
	"context"

	"github.com/superwhys/billiard-helper/dal/db/query"
	"github.com/superwhys/billiard-helper/models/dbmodels"
	"github.com/superwhys/billiard-helper/ports"
	"gorm.io/gorm"
)

type playerManager struct {
	query *query.Query
}

func NewPlayerManager(db *gorm.DB) ports.PlayerRepo {
	return &playerManager{
		query: query.Use(db),
	}
}

func (m *playerManager) CreatePlayer(ctx context.Context, player *dbmodels.Player) error {
	p := m.query.Player
	return p.WithContext(ctx).Create(player)
}

func (m *playerManager) GetPlayer(ctx context.Context, playerCode string) (*dbmodels.Player, error) {
	p := m.query.Player
	return p.WithContext(ctx).
		Where(p.Code.Eq(playerCode)).
		First()
}

func (m *playerManager) GetPlayerByCode(ctx context.Context, code string) (*dbmodels.Player, error) {
	p := m.query.Player
	return p.WithContext(ctx).
		Where(p.Code.Eq(code)).
		First()
}

func (m *playerManager) DeletePlayer(ctx context.Context, playerCode string) error {
	p := m.query.Player
	_, err := p.WithContext(ctx).
		Where(p.Code.Eq(playerCode)).
		Delete()
	return err
}
