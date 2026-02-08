package db

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/infra/db/assembler"
	"github.com/superwhys/billiard-helper/internal/infra/db/models"
	"github.com/superwhys/billiard-helper/internal/infra/db/query"
	"gorm.io/gorm"
)

var _ match.IPlayerRepository = (*PlayerRepo)(nil)

type PlayerRepo struct {
	query             *query.Query
	playerPoAssembler *assembler.PlayerPoAssembler
}

func NewPlayerRepo(db *gorm.DB) *PlayerRepo {
	return &PlayerRepo{query: query.Use(db)}
}

func (r *PlayerRepo) FindByCode(ctx context.Context, code string) (*match.Player, error) {
	p := r.query.Player
	po, err := p.WithContext(ctx).Where(p.Code.Eq(code)).First()
	if err != nil {
		return nil, err
	}
	return r.playerPoAssembler.ToEntity(po), nil
}

func (r *PlayerRepo) Create(ctx context.Context, player *match.Player) error {
	po := r.playerPoAssembler.ToPO(player)
	err := r.query.Player.WithContext(ctx).Create(po)
	if err != nil {
		return err
	}
	return nil
}

func (r *PlayerRepo) CreateInBatches(ctx context.Context, players []*match.Player) error {
	pos := make([]*models.Player, 0, len(players))
	for _, player := range players {
		pos = append(pos, r.playerPoAssembler.ToPO(player))
	}

	batchSize := min(len(players), 30)
	err := r.query.Player.WithContext(ctx).CreateInBatches(pos, batchSize)
	if err != nil {
		return err
	}

	// 回填 playerID
	for i, player := range players {
		player.ID = pos[i].ID
	}
	return nil
}

func (r *PlayerRepo) Delete(ctx context.Context, id uint) error {
	p := r.query.Player
	_, err := p.WithContext(ctx).
		Where(p.ID.Eq(id)).
		Delete()
	if err != nil {
		return err
	}
	return nil
}
