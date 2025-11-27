package dto

import (
	"time"

	"github.com/superwhys/billiard-helper/internal/domain/match"
)

type GameConfig struct {
	GameType    int `json:"game_type"` // 比如 1: snooker, 2: 8-ball
	MaxPlayers  int `json:"max_players"`
	TargetScore int `json:"target_score"`
}

type Match struct {
	ID        uint       `json:"id"`
	OwnerID   uint       `json:"owner_id"`
	Status    int        `json:"status"`
	Config    GameConfig `json:"config"`
	Players   []Player   `json:"players"`
	CreatedAt time.Time  `json:"created_at"`
}

type Player struct {
	ID       uint      `json:"id"`
	Code     string    `json:"code"`
	UserID   *uint     `json:"user_id,omitempty"`
	NickName string    `json:"nick_name"`
	Type     int       `json:"type"`
	JoinTime time.Time `json:"join_time"`
}

type CreateMatchRequest struct {
	Operator
	GameType    int `json:"game_type"`
	MaxPlayers  int `json:"max_players"`
	TargetScore int `json:"target_score"`
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
