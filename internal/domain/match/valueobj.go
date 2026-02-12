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

type MatchType string

const (
	MatchTypeSnooker MatchType = "snooker"
	MatchType8Ball   MatchType = "8ball"
	MatchType9Ball   MatchType = "9ball"
)

// MatchConfig 比赛配置值对象
// 仅存储配置信息，不包含计算逻辑
type MatchConfig struct {
	MaxPlayers  uint           `json:"max_players"`  // 最大玩家数量
	TargetScore uint           `json:"target_score"` // 目标分数（如抢几）
	Data        map[string]any `json:"data"`         // 其他配置数据

}

func MatchTypeMaxPlayers(mt MatchType) uint {
	switch mt {
	case MatchTypeSnooker:
		return 2
	case MatchType9Ball:
		return 5
	case MatchType8Ball:
		return 2
	default:
		return 2
	}
}

func Default9BallScoreConfig() map[string]any {
	return map[string]any{
		"big":    10,
		"small":  7,
		"golden": 4,
		"win":    4,
		"foul":   -1,
	}
}
