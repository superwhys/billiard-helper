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

func (r *MatchRepo) FindByID(ctx context.Context, id uint) (*match.Match, error) {
	m := r.query.Match

	// 需要预加载 Players
	po, err := m.WithContext(ctx).
		Preload(m.Players).
		Where(m.ID.Eq(id)).
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

	// 使用 Save 更新整个聚合根，包括关联的 Players
	// 注意：GORM 的 Save 会更新所有字段，包括零值。对于关联关系，如果配置了 FullSaveAssociations，会保存关联。
	// 在 gen 中，Save 对应的是 gorm.Save
	err := m.WithContext(ctx).Save(po)
	if err != nil {
		return err
	}

	// 更新时间
	match.UpdatedAt = po.UpdatedAt

	// 回写可能新增的 Player IDs
	if len(match.Players) > 0 && len(po.Players) == len(match.Players) {
		for i, p := range match.Players {
			if p.ID == 0 {
				p.ID = po.Players[i].ID
				p.JoinTime = po.Players[i].CreatedAt
			}
		}
	}

	return nil
}

func (r *MatchRepo) Delete(ctx context.Context, id uint) error {
	m := r.query.Match
	_, err := m.WithContext(ctx).Where(m.ID.Eq(id)).Delete()
	return err
}
