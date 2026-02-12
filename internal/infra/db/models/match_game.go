package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// MatchGame 一场比赛内的一局
type MatchGame struct {
	gorm.Model
	MatchID     uint           `gorm:"column:match_id;index;not null;comment:比赛ID" json:"match_id"`
	GameNum     uint           `gorm:"column:game_num;not null;comment:局数" json:"game_num"`
	StartAt     int64          `gorm:"column:start_at;not null;comment:开始时间" json:"start_at"`
	EndAt       int64          `gorm:"column:end_at;not null;comment:结束时间" json:"end_at"`
	Scores      datatypes.JSON `gorm:"column:scores;type:json;comment:玩家分数快照" json:"scores"`
	LastEventID *uint          `gorm:"column:last_event_id;index;comment:最新事件ID" json:"last_event_id"`
	WinnerID    *uint          `gorm:"column:winner_id;index;comment:赢家ID" json:"winner_id"`

	Winner    *Player     `gorm:"foreignKey:WinnerID" json:"winner"`
	LastEvent *MatchEvent `gorm:"foreignKey:LastEventID" json:"last_event"`
}

func (gm *MatchGame) TableName() string {
	return "match_games"
}
