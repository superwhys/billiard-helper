package match

import (
	"fmt"
	"time"

	"github.com/miebyte/goutils/utils/ptrx"
	"github.com/superwhys/billiard-helper/internal/errcode"
	"github.com/superwhys/billiard-helper/internal/pkg/codegen"
)

// Match 聚合根
type Match struct {
	ID        uint        `json:"id"`
	OwnerID   uint        `json:"owner_id"`
	Status    MatchStatus `json:"status"`
	MatchType MatchType   `json:"match_type"`
	Config    MatchConfig `json:"config"`
	Players   []*Player   `json:"players"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

func (m *Match) IsPlayerFull() bool {
	if len(m.Players) >= int(m.Config.MaxPlayers) {
		return true
	}

	return false
}

func (m *Match) IsPlayerOutOfLimit() bool {
	if len(m.Players) > int(m.Config.MaxPlayers) {
		return true
	}

	return false
}

func (m *Match) JoinPlayer(player *Player) error {
	if m.Status != MatchStatusPending {
		return errcode.ErrCodeMatchNotPending
	}

	if m.IsPlayerFull() {
		return errcode.ErrCodeMatchPlayerFull
	}

	// 检查是否重复加入
	for _, p := range m.Players {
		// 如果是真实玩家，检查 id 是否相同
		if player.Type == PlayerTypeReal {
			if p.UserID != nil && *p.UserID == *player.UserID {
				return errcode.ErrCodeMatchPlayerAlreadyJoined
			} else {
				// 如果是虚拟玩家，检查昵称是否相同
				if p.NickName == player.NickName {
					return errcode.ErrCodeMatchPlayerAlreadyJoined
				}
			}
		}
	}

	m.Players = append(m.Players, player)
	return nil
}

type Player struct {
	ID       uint       `json:"id"`
	Code     string     `json:"code"`
	MatchID  uint       `json:"match_id"`
	UserID   *uint      `json:"user_id"`
	NickName string     `json:"nick_name"`
	Type     PlayerType `json:"type"`
	JoinTime time.Time  `json:"join_time"`
}

func NewPlayer(matchID uint, userID *uint, nickName string, pType PlayerType) *Player {
	p := &Player{
		MatchID:  matchID,
		UserID:   userID,
		NickName: nickName,
		Type:     pType,
		JoinTime: time.Now(),
	}

	p.GenerateCode(matchID, userID, nickName)
	return p
}

func (p *Player) GenerateCode(matchID uint, userID *uint, nickName string) {
	var payload string
	if p.Type == PlayerTypeReal {
		payload = fmt.Sprintf("%d", ptrx.UintValue(p.UserID))
	} else {
		payload = p.NickName
	}

	p.Code = codegen.GeneratePlayerCode(matchID, uint8(p.Type), payload)
}
