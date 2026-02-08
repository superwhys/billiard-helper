package models

import "gorm.io/gorm"

// Player 玩家模型
// 一个比赛可以有多个玩家，一个玩家可以参加多个比赛
// 可以支持虚拟玩家和真实玩家
// 虚拟玩家没有用户ID，真实玩家有用户ID
type Player struct {
	gorm.Model
	MatchID  uint   `gorm:"column:match_id;uniqueIndex:idx_match_id_code;not null;comment:比赛ID" json:"match_id"`
	Code     string `gorm:"column:code;type:varchar(255);index;uniqueIndex:idx_match_id_code;not null;comment:玩家代码" json:"code"`
	UserID   *uint  `gorm:"column:user_id;index;comment:真实玩家的用户ID" json:"user_id"`
	NickName string `gorm:"column:nick_name;type:varchar(255);not null;comment:昵称" json:"nick_name"`
	Type     uint8  `gorm:"column:type;type:tinyint(1);default:1;not null;comment:玩家类型" json:"type"`

	Events []*MatchEvent `json:"events"`
}

func (p *Player) TableName() string {
	return "players"
}
