package dbmodels

import (
	"github.com/superwhys/billiard-helper/models/types"
	"gorm.io/gorm"
)

type Player struct {
	gorm.Model
	RoomID    uint             `gorm:"column:room_id;index;not null;comment:房间ID" json:"room_id"`
	UserID    uint             `gorm:"column:user_id;index;comment:用户ID" json:"user_id"`
	NickName  string           `gorm:"column:nick_name;type:varchar(255);not null;comment:昵称" json:"nick_name"`
	AvatarURL string           `gorm:"column:avatar_url;type:varchar(255);comment:头像URL" json:"avatar_url"`
	Type      types.PlayerType `gorm:"column:type;type:tinyint(1);default:1;not null;comment:玩家类型" json:"type"`

	Scores []*Scores `json:"scores"`
}

func (p *Player) TableName() string {
	return "players"
}
