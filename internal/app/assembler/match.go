package assembler

import (
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/domain/match"
)

type MatchAssembler struct{}

func NewMatchAssembler() *MatchAssembler {
	return &MatchAssembler{}
}

func (a *MatchAssembler) CreateMatchReqToMatch(req *dto.CreateMatchRequest) *match.Match {
	if req == nil {
		return nil
	}

	m := &match.Match{
		OwnerID:   req.UserID,
		Status:    match.MatchStatusPending,
		MatchType: req.MatchType,
		Config: match.MatchConfig{
			MaxPlayers:  match.MatchTypeMaxPlayers(req.MatchType),
			TargetScore: uint(req.TargetScore),
		},
		Players: make([]*match.Player, 0, len(req.VirtualPlayers)),
	}

	for _, dtoPlayer := range req.VirtualPlayers {
		m.Players = append(m.Players, a.DtoToPlayer(dtoPlayer))
	}

	return m
}

func (a *MatchAssembler) ToMatchDTO(r *match.Match) *dto.Match {
	if r == nil {
		return nil
	}

	players := make([]dto.Player, 0, len(r.Players))
	for _, p := range r.Players {
		players = append(players, a.ToPlayerDTO(p))
	}

	config := dto.MatchConfig{
		MaxPlayers:  r.Config.MaxPlayers,
		TargetScore: r.Config.TargetScore,
	}

	return &dto.Match{
		ID:        r.ID,
		OwnerID:   r.OwnerID,
		Status:    int(r.Status),
		Config:    config,
		Players:   players,
		CreatedAt: r.CreatedAt,
	}
}

func (a *MatchAssembler) DtoToPlayer(p *dto.Player) *match.Player {
	if p == nil {
		return nil
	}

	return &match.Player{
		Code:     p.Code,
		UserID:   p.UserID,
		NickName: p.NickName,
		Type:     p.Type,
		JoinTime: p.JoinTime,
	}
}

func (a *MatchAssembler) ToPlayerDTO(p *match.Player) dto.Player {
	if p == nil {
		return dto.Player{}
	}
	return dto.Player{
		ID:       p.ID,
		Code:     p.Code,
		UserID:   p.UserID,
		NickName: p.NickName,
		Type:     p.Type,
		JoinTime: p.JoinTime,
	}
}

// ToMatchConfig 将 DTO 配置转换为领域值对象
func (a *MatchAssembler) ToMatchConfig(req *dto.CreateMatchRequest) match.MatchConfig {
	return match.MatchConfig{
		MaxPlayers:  uint(req.MaxPlayers),
		TargetScore: uint(req.TargetScore),
	}
}

func (a *MatchAssembler) ToDtoMatchConfig(config *match.MatchConfig) dto.MatchConfig {
	return dto.MatchConfig{
		MaxPlayers:  config.MaxPlayers,
		TargetScore: config.TargetScore,
	}
}
