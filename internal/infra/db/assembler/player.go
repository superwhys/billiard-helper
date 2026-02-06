package assembler

import (
	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/infra/db/models"
	"gorm.io/gorm"
)

type PlayerPoAssembler struct{}

func NewPlayerPoAssembler() *PlayerPoAssembler {
	return &PlayerPoAssembler{}
}

func (a *PlayerPoAssembler) ToEntity(po *models.Player) *match.Player {
	if po == nil {
		return nil
	}

	return &match.Player{
		ID:       po.ID,
		Code:     po.Code,
		MatchID:  po.MatchID,
		UserID:   po.UserID,
		NickName: po.NickName,
		Type:     match.PlayerType(po.Type),
		JoinTime: po.CreatedAt,
	}
}

func (a *PlayerPoAssembler) ToPO(entity *match.Player) *models.Player {
	if entity == nil {
		return nil
	}

	return &models.Player{
		Model: gorm.Model{
			ID: entity.ID,
		},
		Code:     entity.Code,
		MatchID:  entity.MatchID,
		UserID:   entity.UserID,
		NickName: entity.NickName,
		Type:     uint8(entity.Type),
	}
}
