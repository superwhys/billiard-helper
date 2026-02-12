package db

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/domain/event"
	"github.com/superwhys/billiard-helper/internal/infra/db/assembler"
	"github.com/superwhys/billiard-helper/internal/infra/db/query"
	"gorm.io/gorm"
)

var _ event.IEventRepository = (*MatchEventRepo)(nil)

type MatchEventRepo struct {
	query                 *query.Query
	matchEventPoAssembler *assembler.MatchEventPoAssembler
}

func NewMatchEventRepo(db *gorm.DB) *MatchEventRepo {
	return &MatchEventRepo{
		query:                 query.Use(db),
		matchEventPoAssembler: assembler.NewMatchEventPoAssembler(),
	}
}

func (r *MatchEventRepo) AddEvent(ctx context.Context, event *event.Event) error {
	me := r.query.MatchEvent
	po := r.matchEventPoAssembler.ToPO(event)

	err := me.WithContext(ctx).Create(po)
	if err != nil {
		return err
	}

	event.ID = po.ID
	return nil
}

func (r *MatchEventRepo) GetMatchEvents(ctx context.Context, matchID uint) ([]*event.Event, error) {
	me := r.query.MatchEvent
	pos, err := me.WithContext(ctx).Where(me.MatchID.Eq(matchID)).Find()
	if err != nil {
		return nil, err
	}
	return r.matchEventPoAssembler.ToEntityList(pos), nil
}

func (r *MatchEventRepo) DeleteEvent(ctx context.Context, eventID uint) (*event.Event, error) {
	me := r.query.MatchEvent
	po, err := me.WithContext(ctx).Where(me.ID.Eq(eventID)).First()
	if err != nil {
		return nil, err
	}
	return r.matchEventPoAssembler.ToEntity(po), nil
}

func (r *MatchEventRepo) GetLastEvent(ctx context.Context, matchID uint, round uint) (*event.Event, error) {
	me := r.query.MatchEvent
	po, err := me.WithContext(ctx).
		Where(
			me.MatchID.Eq(matchID),
			me.Round.Eq(round),
		).
		Order(me.Round.Desc()).
		First()
	if err != nil {
		return nil, err
	}
	return r.matchEventPoAssembler.ToEntity(po), nil
}
