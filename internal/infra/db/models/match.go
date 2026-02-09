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
	UserID     uint                            `gorm:"column:user_id;index;not null;comment:房主ID" json:"user_id"`
	Name       string                          `gorm:"column:name;type:varchar(255);not null;comment:比赛名称" json:"name"`
	Status     uint8                           `gorm:"column:status;type:tinyint(1);default:1;not null;comment:房间状态" json:"status"`
	MatchType  string                          `gorm:"column:match_type;type:varchar(255);not null;comment:比赛类型" json:"match_type"`
	MatchRound uint                            `gorm:"column:match_round;type:int(11);default:1;not null;comment:比赛轮数(正在进行第几轮)" json:"match_round"`
	Config     datatypes.JSONType[MatchConfig] `gorm:"column:config;type:json;comment:比赛配置" json:"config"`

	Players    []*Player     `json:"players"`
	Events     []*MatchEvent `json:"events"`
	MatchGames []*MatchGame  `json:"match_games"`
}

func (r *Match) TableName() string {
	return "matches"
}
