package assembler

import (
	"encoding/json"

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
		MaxPlayers:  po.MaxPlayers,
		TargetScore: po.TargetScore,
		Data:        po.Data,
	}
}

func (a *MatchPoAssembler) ToMatchConfigEntity(config models.MatchConfig) match.MatchConfig {
	return match.MatchConfig{
		MaxPlayers:  config.MaxPlayers,
		TargetScore: config.TargetScore,
		Data:        config.Data,
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

	matchGames := make([]*match.MatchGame, 0, len(po.MatchGames))
	for _, m := range po.MatchGames {
		matchGames = append(matchGames, a.ToMatchGameEntity(m))
	}

	config := a.ToMatchConfigEntity(po.Config.Data())

	return &match.Match{
		ID:         po.ID,
		OwnerID:    po.UserID,
		Name:       po.Name,
		Status:     match.MatchStatus(po.Status),
		Config:     config,
		MatchType:  match.MatchType(po.MatchType),
		MatchRound: po.MatchRound,
		Players:    players,
		MatchGames: matchGames,
		CreatedAt:  po.CreatedAt,
		UpdatedAt:  po.UpdatedAt,
	}
}

func (a *MatchPoAssembler) ToEntityList(pos []*models.Match) []*match.Match {
	lst := make([]*match.Match, 0, len(pos))
	for _, po := range pos {
		lst = append(lst, a.ToEntity(po))
	}

	return lst
}

func (a *MatchPoAssembler) ToPO(entity *match.Match) *models.Match {
	if entity == nil {
		return nil
	}

	config := a.ToMatchConfig(entity.Config)
	m := &models.Match{
		Model: gorm.Model{
			ID: entity.ID,
		},
		UserID:     entity.OwnerID,
		Status:     uint8(entity.Status),
		MatchRound: entity.MatchRound,
		Name:       entity.Name,
		Config:     datatypes.NewJSONType(config),
		MatchType:  string(entity.MatchType),
	}

	return m
}

func (a *MatchPoAssembler) ToMatchGameEntity(mg *models.MatchGame) *match.MatchGame {
	if mg == nil {
		return nil
	}

	return &match.MatchGame{
		ID:          mg.ID,
		MatchID:     mg.MatchID,
		GameNum:     mg.GameNum,
		StartAt:     mg.StartAt,
		EndAt:       mg.EndAt,
		Scores:      json.RawMessage(mg.Scores),
		LastEventID: mg.LastEventID,
		WinnerID:    mg.WinnerID,
	}
}
