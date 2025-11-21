package dbmodels

import (
	"github.com/superwhys/billiard-helper/internal/models/types"
	"gorm.io/gorm"
)

type Scores struct {
	gorm.Model
	PlayerID   uint `gorm:"column:player_id;index;not null;comment:操作分数的玩家ID" json:"player_id"`
	RoomID     uint `gorm:"column:room_id;index;not null;comment:操作分数的房间ID" json:"room_id"`
	OperatorID uint `gorm:"column:operator_id;index;not null;comment:操作人ID" json:"operator_id"`
	Change     int  `gorm:"column:change;not null;comment:变化值" json:"change"`
	Total      int  `gorm:"column:total;not null;comment:总分" json:"total"`

	Operator *Player `gorm:"foreignKey:PlayerID" json:"operator"`
}

func (s *Scores) TableName() string {
	return "scores"
}

func (s *Scores) ToType() *types.Scores {
	scoresT := &types.Scores{
		PlayerID:   s.PlayerID,
		RoomID:     s.RoomID,
		OperatorID: s.OperatorID,
		Change:     s.Change,
		Total:      s.Total,
	}

	if s.Operator != nil {
		scoresT.Operator = s.Operator.ToType()
	}

	return scoresT
}
