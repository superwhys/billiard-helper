package assembler

import (
	"github.com/superwhys/billiard-helper/internal/app/dto"
	"github.com/superwhys/billiard-helper/internal/domain/match"
)

type MatchAssembler struct{}

func NewMatchAssembler() *MatchAssembler {
	return &MatchAssembler{}
}

// ToRoomDTO 将 Room 聚合根转换为 DTO
func (a *MatchAssembler) ToMatchDTO(r *match.Match) *dto.Match {
	if r == nil {
		return nil
	}

	players := make([]dto.Player, 0, len(r.Players))
	for _, p := range r.Players {
		players = append(players, a.ToPlayerDTO(p))
	}

	config := dto.MatchConfig{
		MatchType:   uint8(r.Config.MatchType),
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

// ToPlayerDTO 将 Player 实体转换为 DTO
func (a *MatchAssembler) ToPlayerDTO(p *match.Player) dto.Player {
	if p == nil {
		return dto.Player{}
	}
	return dto.Player{
		ID:       p.ID,
		Code:     p.Code,
		UserID:   p.UserID,
		NickName: p.NickName,
		Type:     int(p.Type),
		JoinTime: p.JoinTime,
	}
}

// ToMatchConfig 将 DTO 配置转换为领域值对象
func (a *MatchAssembler) ToMatchConfig(req *dto.CreateMatchRequest) match.MatchConfig {
	return match.MatchConfig{
		MatchType:   match.MatchType(req.MatchType),
		MaxPlayers:  uint(req.MaxPlayers),
		TargetScore: uint(req.TargetScore),
	}
}

func (a *MatchAssembler) ToDtoMatchConfig(config *match.MatchConfig) dto.MatchConfig {
	return dto.MatchConfig{
		MatchType:   uint8(config.MatchType),
		MaxPlayers:  config.MaxPlayers,
		TargetScore: config.TargetScore,
	}
}
