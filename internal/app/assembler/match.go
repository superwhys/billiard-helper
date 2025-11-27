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
func (a *MatchAssembler) ToRoomDTO(r *match.Room) *dto.Room {
	if r == nil {
		return nil
	}

	players := make([]dto.Player, 0, len(r.Players))
	for _, p := range r.Players {
		players = append(players, a.ToPlayerDTO(p))
	}

	config := dto.GameConfig{
		GameType:    int(r.Config.GameType),
		MaxPlayers:  r.Config.MaxPlayers,
		TargetScore: r.Config.TargetScore,
	}

	return &dto.Room{
		ID:        r.ID,
		RoomCode:  r.RoomCode,
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
		UserID:   p.UserID,
		NickName: p.NickName,
		Type:     int(p.Type),
		IsOnline: p.IsOnline,
		JoinTime: p.JoinTime,
	}
}

// ToGameConfig 将 DTO 配置转换为领域值对象
func (a *MatchAssembler) ToGameConfig(req *dto.CreateRoomRequest) match.GameConfig {
	return match.GameConfig{
		GameType:    match.GameType(req.GameType),
		MaxPlayers:  req.MaxPlayers,
		TargetScore: req.TargetScore,
	}
}

