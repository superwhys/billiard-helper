package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type MatchConfig struct {
	MatchType   uint8 `json:"match_type"`   // 房间类型(台球类型)
	MaxPlayers  uint  `json:"max_players"`  // 最大玩家数量
	TargetScore uint  `json:"target_score"` // 目标分数（如抢几）
}

type Match struct {
	gorm.Model
	UserID uint                            `gorm:"column:user_id;index;not null;comment:房主ID" json:"user_id"`
	Status uint8                           `gorm:"column:status;type:tinyint(1);default:1;not null;comment:房间状态" json:"status"`
	Config datatypes.JSONType[MatchConfig] `gorm:"column:config;type:json;comment:比赛配置" json:"config"`

	Players []*Player `json:"players"`
	Scores  []*Score  `json:"scores"`
}

func (r *Match) TableName() string {
	return "matches"
}
