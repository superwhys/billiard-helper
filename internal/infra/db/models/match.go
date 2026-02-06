package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type MatchConfig struct {
	MaxPlayers  uint `json:"max_players"`  // 最大玩家数量
	TargetScore uint `json:"target_score"` // 目标分数（如抢几）
}

type Match struct {
	gorm.Model
	UserID    uint                            `gorm:"column:user_id;index;not null;comment:房主ID" json:"user_id"`
	Status    uint8                           `gorm:"column:status;type:tinyint(1);default:1;not null;comment:房间状态" json:"status"`
	MatchType string                          `gorm:"column:match_type;type:varchar(255);not null;comment:比赛类型" json:"match_type"`
	Config    datatypes.JSONType[MatchConfig] `gorm:"column:config;type:json;comment:比赛配置" json:"config"`

	Players    []*Player    `json:"players"`
	Events     []*Event     `json:"events"`
	MatchGames []*MatchGame `json:"match_games"`
}

func (r *Match) TableName() string {
	return "matches"
}
