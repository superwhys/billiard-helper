package dto

import (
	"time"

	"github.com/superwhys/billiard-helper/internal/domain/match"
)

type MatchConfig struct {
	MaxPlayers  uint `json:"max_players"`
	TargetScore uint `json:"target_score"`
}

type Match struct {
	ID         uint            `json:"id"`
	OwnerID    uint            `json:"owner_id"`
	Name       string          `json:"name"`
	Status     int             `json:"status"`
	MatchType  match.MatchType `json:"match_type"`
	MatchRound uint            `json:"match_round"`
	Config     MatchConfig     `json:"config"`
	Players    []Player        `json:"players"`
	CreatedAt  time.Time       `json:"created_at"`
}

type Player struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	// 真实用户的用户 id
	UserID   *uint            `json:"user_id,omitempty"`
	NickName string           `json:"nick_name"`
	Type     match.PlayerType `json:"type"`
	JoinTime time.Time        `json:"join_time"`
}

type CreateMatchRequest struct {
	Operator
	Name           string          `json:"name"`
	MatchType      match.MatchType `json:"match_type"`
	MaxPlayers     uint            `json:"max_players"`
	TargetScore    uint            `json:"target_score"`
	VirtualPlayers []*Player       `json:"virtual_players"`
}

type UpdateMatchRequest struct {
	Operator
	MatchID     uint   `json:"match_id"`
	Name        string `json:"name"`
	TargetScore uint   `json:"target_score"`
}

type JoinMatchRequest struct {
	Operator
	MatchID    uint             `json:"match_id"`
	NickName   string           `json:"nick_name"`
	PlayerType match.PlayerType `json:"player_type"`
}

type MatchActionRequest struct {
	Operator
	MatchID    uint   `json:"match_id"`
	PlayerCode string `json:"player_code"`
}

type KickPlayerRequest struct {
	Operator
	MatchID    uint   `json:"match_id"`
	PlayerCode string `json:"player_code"`
}

type MatchListRequest struct {
	MatchType match.MatchType `form:"match_type"`
	Limit     uint            `form:"limit"`
	Cursor    uint            `form:"cursor"`
}

type MatchDetailRequest struct {
	MatchID uint `form:"match_id"`
}
