package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/infra/db/assembler"
	"github.com/superwhys/billiard-helper/internal/infra/db/models"
	"github.com/superwhys/billiard-helper/internal/infra/db/query"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var _ match.IMatchGameRepository = (*MatchGameRepo)(nil)

type MatchGameRepo struct {
	query                *query.Query
	matchGamePoAssembler *assembler.MatchGamePoAssembler
}

func NewMatchGameRepo(db *gorm.DB) *MatchGameRepo {
	return &MatchGameRepo{
		query:                query.Use(db),
		matchGamePoAssembler: assembler.NewMatchGamePoAssembler(),
	}
}

func (r *MatchGameRepo) Create(ctx context.Context, matchGame *match.MatchGame) error {
	mg := r.query.MatchGame
	po := r.matchGamePoAssembler.ToPO(matchGame)

	err := mg.WithContext(ctx).Create(po)
	if err != nil {
		return err
	}

	// 回写 ID
	matchGame.ID = po.ID

	return nil
}

func (r *MatchGameRepo) FindByMatchID(ctx context.Context, matchID uint, gameNum uint) (*match.MatchGame, error) {
	mg := r.query.MatchGame
	po, err := mg.WithContext(ctx).
		Where(
			mg.MatchID.Eq(matchID),
			mg.GameNum.Eq(gameNum),
		).
		First()
	if err != nil {
		return nil, err
	}

	return r.matchGamePoAssembler.ToEntity(po), nil
}

func (r *MatchGameRepo) FindMatchRounds(ctx context.Context, matchID uint) ([]*match.MatchGame, error) {
	mg := r.query.MatchGame
	pos, err := mg.WithContext(ctx).
		Where(mg.MatchID.Eq(matchID)).
		Order(mg.GameNum.Desc()).
		Find()
	if err != nil {
		return nil, err
	}
	return r.matchGamePoAssembler.ToEntityList(pos), nil
}

func (r *MatchGameRepo) EndGameRound(ctx context.Context, matchID uint, gameNum uint) error {
	mg := r.query.MatchGame
	_, err := mg.WithContext(ctx).
		Where(
			mg.MatchID.Eq(matchID),
			mg.GameNum.Eq(gameNum),
		).
		Update(mg.EndAt, time.Now().Unix())
	if err != nil {
		return err
	}
	return nil
}

func (r *MatchGameRepo) StartGameRound(ctx context.Context, matchID uint, gameNum uint, scores json.RawMessage) (*match.MatchGame, error) {
	mg := r.query.MatchGame

	po := &models.MatchGame{
		MatchID: matchID,
		GameNum: gameNum,
		StartAt: time.Now().Unix(),
		Scores:  datatypes.JSON(scores),
	}
	err := mg.WithContext(ctx).Create(po)
	if err != nil {
		return nil, fmt.Errorf("create match game failed: %w", err)
	}
	return r.matchGamePoAssembler.ToEntity(po), nil
}

func (r *MatchGameRepo) Update(ctx context.Context, matchGame *match.MatchGame) error {
	mg := r.query.MatchGame
	po := r.matchGamePoAssembler.ToPO(matchGame)

	_, err := mg.WithContext(ctx).
		Where(mg.ID.Eq(matchGame.ID)).
		Updates(po)
	if err != nil {
		return err
	}
	return nil
}

func (r *MatchGameRepo) Delete(ctx context.Context, id uint) error {
	mg := r.query.MatchGame
	_, err := mg.WithContext(ctx).
		Where(mg.ID.Eq(id)).
		Delete()
	if err != nil {
		return err
	}
	return nil
}
