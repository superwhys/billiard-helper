package dbmodels

import (
	"github.com/miebyte/goutils/utils/ptrx"
	"github.com/superwhys/billiard-helper/internal/models/types"
	"gorm.io/gorm"
)

type Player struct {
	gorm.Model
	Code     string           `gorm:"column:code;type:varchar(255);unique;not null;comment:玩家代码" json:"code"`
	RoomID   uint             `gorm:"column:room_id;index;not null;uniqueIndex:idx_room_id_user_id;comment:房间ID" json:"room_id"`
	UserID   *uint            `gorm:"column:user_id;index;uniqueIndex:idx_room_id_user_id;comment:用户ID" json:"user_id"`
	NickName string           `gorm:"column:nick_name;type:varchar(255);not null;comment:昵称" json:"nick_name"`
	Type     types.PlayerType `gorm:"column:type;type:tinyint(1);default:1;not null;comment:玩家类型" json:"type"`
	IsOnline bool             `gorm:"column:is_online;type:tinyint(1);default:1;not null;comment:是否在线" json:"is_online"`

	Scores []*Scores `json:"scores"`
}

func (p *Player) TableName() string {
	return "players"
}

func (p *Player) ToType() *types.Player {
	playerT := &types.Player{
		ID:       p.ID,
		Code:     p.Code,
		RoomID:   p.RoomID,
		UserID:   ptrx.UintValue(p.UserID),
		NickName: p.NickName,
		Type:     p.Type,
		IsOnline: p.IsOnline,
	}
	return playerT
}
