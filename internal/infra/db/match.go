package db

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/infra/db/assembler"
	"github.com/superwhys/billiard-helper/internal/infra/db/query"
	"gorm.io/gorm"
)

var _ match.IMatchRepository = (*MatchRepo)(nil)

type MatchRepo struct {
	query            *query.Query
	matchPoAssembler *assembler.MatchPoAssembler
}

func NewMatchRepo(db *gorm.DB) *MatchRepo {
	return &MatchRepo{
		query:            query.Use(db),
		matchPoAssembler: assembler.NewMatchPoAssembler(),
	}
}

func (r *MatchRepo) Create(ctx context.Context, match *match.Match) error {
	m := r.query.Match
	po := r.matchPoAssembler.ToPO(match)

	err := m.WithContext(ctx).Create(po)
	if err != nil {
		return err
	}

	// 回写 ID
	match.ID = po.ID
	match.CreatedAt = po.CreatedAt
	match.UpdatedAt = po.UpdatedAt

	// 回写 Player IDs
	if len(match.Players) > 0 && len(po.Players) == len(match.Players) {
		for i, p := range match.Players {
			p.ID = po.Players[i].ID
			p.JoinTime = po.Players[i].CreatedAt
		}
	}

	return nil
}

func (r *MatchRepo) FindByID(ctx context.Context, id uint, withPlayers bool) (*match.Match, error) {
	m := r.query.Match

	query := m.WithContext(ctx).Where(m.ID.Eq(id))
	if withPlayers {
		query = query.Preload(m.Players)
	}

	po, err := query.First()
	if err != nil {
		return nil, err
	}

	return r.matchPoAssembler.ToEntity(po), nil
}

func (r *MatchRepo) GetMatchDetail(ctx context.Context, id uint) (*match.Match, error) {
	m := r.query.Match

	po, err := m.WithContext(ctx).
		Where(m.ID.Eq(id)).
		Preload(m.Players).
		Preload(m.MatchGames).
		First()

	if err != nil {
		return nil, err
	}
	return r.matchPoAssembler.ToEntity(po), nil
}

func (r *MatchRepo) IsExists(ctx context.Context, id uint) (bool, error) {
	m := r.query.Match
	count, err := m.WithContext(ctx).Where(m.ID.Eq(id)).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *MatchRepo) Update(ctx context.Context, match *match.Match) error {
	m := r.query.Match
	po := r.matchPoAssembler.ToPO(match)

	_, err := m.WithContext(ctx).
		Where(m.ID.Eq(match.ID)).
		Updates(po)
	if err != nil {
		return err
	}

	// 更新时间
	match.UpdatedAt = po.UpdatedAt
	return nil
}

func (r *MatchRepo) Delete(ctx context.Context, id uint) error {
	m := r.query.Match
	_, err := m.WithContext(ctx).Where(m.ID.Eq(id)).Delete()
	return err
}

func (r *MatchRepo) ListMatches(ctx context.Context, userID uint, matchType string, limit uint, cursor uint) ([]*match.Match, error) {
	m := r.query.Match

	query := r.query.Match.WithContext(ctx)
	if matchType != "" {
		query = query.Where(m.MatchType.Eq(matchType))
	}

	if cursor > 0 {
		query = query.Where(m.ID.Lt(cursor))
	}

	matches, err := query.
		Where(m.UserID.Eq(userID)).
		Preload(m.Players).
		Preload(m.MatchGames).
		Limit(int(limit)).
		Order(m.ID.Desc()).
		Find()
	if err != nil {
		return nil, err
	}

	return r.matchPoAssembler.ToEntityList(matches), nil
}
