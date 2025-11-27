package match

type MatchStatus uint8

const (
	MatchStatusPending    MatchStatus = 1 // 未开始
	MatchStatusInProgress MatchStatus = 2 // 进行中
	MatchStatusFinished   MatchStatus = 3 // 已完成
)

type PlayerType uint8

const (
	PlayerTypeVirtual PlayerType = 1
	PlayerTypeReal    PlayerType = 2
)

type MatchType uint8

const (
	MatchTypeSnooker MatchType = 1
	MatchType8Ball   MatchType = 2
	MatchType9Ball   MatchType = 3
)

// MatchConfig 比赛配置值对象
// 仅存储配置信息，不包含计算逻辑
type MatchConfig struct {
	MatchType   MatchType `json:"match_type"`   // 比赛类型
	MaxPlayers  uint      `json:"max_players"`  // 最大玩家数量
	TargetScore uint      `json:"target_score"` // 目标分数（如抢几）
}

var DefaultMatchConfig = MatchConfig{
	MatchType:  MatchTypeSnooker,
	MaxPlayers: 2,
}
