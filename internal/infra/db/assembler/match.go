package assembler

import (
	"github.com/superwhys/billiard-helper/internal/domain/match"
	"github.com/superwhys/billiard-helper/internal/infra/db/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type MatchPoAssembler struct{}

func NewMatchPoAssembler() *MatchPoAssembler {
	return &MatchPoAssembler{}
}

func (a *MatchPoAssembler) ToMatchConfig(po match.MatchConfig) models.MatchConfig {
	return models.MatchConfig{
		MatchType:   uint8(po.MatchType),
		MaxPlayers:  po.MaxPlayers,
		TargetScore: po.TargetScore,
	}
}

func (a *MatchPoAssembler) ToMatchConfigEntity(config models.MatchConfig) match.MatchConfig {
	return match.MatchConfig{
		MatchType:   match.MatchType(config.MatchType),
		MaxPlayers:  config.MaxPlayers,
		TargetScore: config.TargetScore,
	}
}

func (a *MatchPoAssembler) ToPlayerEntity(po *models.Player) *match.Player {
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

func (a *MatchPoAssembler) ToPlayerPO(entity *match.Player) *models.Player {
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

func (a *MatchPoAssembler) ToEntity(po *models.Match) *match.Match {
	if po == nil {
		return nil
	}

	players := make([]*match.Player, 0, len(po.Players))
	for _, p := range po.Players {
		players = append(players, a.ToPlayerEntity(p))
	}

	config := a.ToMatchConfigEntity(po.Config.Data())

	return &match.Match{
		ID:        po.ID,
		OwnerID:   po.UserID,
		Status:    match.MatchStatus(po.Status),
		Config:    config,
		Players:   players,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}

func (a *MatchPoAssembler) ToPO(entity *match.Match) *models.Match {
	if entity == nil {
		return nil
	}

	players := make([]*models.Player, 0, len(entity.Players))
	for _, p := range entity.Players {
		players = append(players, a.ToPlayerPO(p))
	}

	config := a.ToMatchConfig(entity.Config)
	m := &models.Match{
		Model: gorm.Model{
			ID: entity.ID,
		},
		UserID:  entity.OwnerID,
		Status:  uint8(entity.Status),
		Config:  datatypes.NewJSONType(config),
		Players: players,
	}

	return m
}
