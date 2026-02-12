package assembler

import (
	"encoding/json"

	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/infra/db/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type MatchGamePoAssembler struct{}

func NewMatchGamePoAssembler() *MatchGamePoAssembler {
	return &MatchGamePoAssembler{}
}

func (a *MatchGamePoAssembler) ToEntity(po *models.MatchGame) *match.MatchGame {
	if po == nil {
		return nil
	}

	return &match.MatchGame{
		ID:          po.ID,
		MatchID:     po.MatchID,
		GameNum:     po.GameNum,
		StartAt:     po.StartAt,
		EndAt:       po.EndAt,
		LastEventID: po.LastEventID,
		WinnerID:    po.WinnerID,
		Scores:      json.RawMessage(po.Scores),
	}
}

func (a *MatchGamePoAssembler) ToEntityList(pos []*models.MatchGame) []*match.MatchGame {
	if len(pos) == 0 {
		return nil
	}

	entities := make([]*match.MatchGame, 0, len(pos))
	for _, po := range pos {
		entities = append(entities, a.ToEntity(po))
	}
	return entities
}

func (a *MatchGamePoAssembler) ToPO(entity *match.MatchGame) *models.MatchGame {
	if entity == nil {
		return nil
	}

	return &models.MatchGame{
		Model: gorm.Model{
			ID: entity.ID,
		},
		MatchID:     entity.MatchID,
		GameNum:     entity.GameNum,
		StartAt:     entity.StartAt,
		EndAt:       entity.EndAt,
		Scores:      datatypes.JSON(entity.Scores),
		LastEventID: entity.LastEventID,
		WinnerID:    entity.WinnerID,
	}
}
