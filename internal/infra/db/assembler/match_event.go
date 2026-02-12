package assembler

import (
	"encoding/json"

	"github.com/superwhys/billiard-helper/internal/domain/event"
	"github.com/superwhys/billiard-helper/internal/infra/db/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type MatchEventPoAssembler struct{}

func NewMatchEventPoAssembler() *MatchEventPoAssembler {
	return &MatchEventPoAssembler{}
}

func (a *MatchEventPoAssembler) ToEntity(po *models.MatchEvent) *event.Event {
	if po == nil {
		return nil
	}

	return &event.Event{
		ID:         po.ID,
		MatchID:    po.MatchID,
		Round:      po.Round,
		OperatorID: po.OperatorID,
		EventType:  po.EventType,
		Data:       json.RawMessage(po.Data),
	}
}

func (a *MatchEventPoAssembler) ToEntityList(pos []*models.MatchEvent) []*event.Event {
	if len(pos) == 0 {
		return nil
	}

	entities := make([]*event.Event, 0, len(pos))
	for _, po := range pos {
		entities = append(entities, a.ToEntity(po))
	}
	return entities
}

func (a *MatchEventPoAssembler) ToPO(entity *event.Event) *models.MatchEvent {
	if entity == nil {
		return nil
	}

	return &models.MatchEvent{
		Model: gorm.Model{
			ID: entity.ID,
		},
		MatchID:    entity.MatchID,
		Round:      entity.Round,
		OperatorID: entity.OperatorID,
		EventType:  entity.EventType,
		Data:       datatypes.JSON(entity.Data),
	}
}
