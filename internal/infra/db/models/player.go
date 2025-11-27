package models

import "gorm.io/gorm"

type Player struct {
	gorm.Model
	Code     string `gorm:"column:code;type:varchar(255);unique;not null;comment:玩家代码" json:"code"`
	MatchID  uint   `gorm:"column:match_id;index;not null;uniqueIndex:idx_match_id_user_id;comment:比赛ID" json:"match_id"`
	UserID   *uint  `gorm:"column:user_id;index;uniqueIndex:idx_match_id_user_id;comment:用户ID" json:"user_id"`
	NickName string `gorm:"column:nick_name;type:varchar(255);not null;comment:昵称" json:"nick_name"`
	Type     uint8  `gorm:"column:type;type:tinyint(1);default:1;not null;comment:玩家类型" json:"type"`

	Scores []*Score `json:"scores"`
}

func (p *Player) TableName() string {
	return "players"
}
