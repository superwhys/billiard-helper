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

type GameType uint8

const (
	GameTypeSnooker GameType = 1
	GameType8Ball   GameType = 2
	GameType9Ball   GameType = 3
)

// GameConfig 比赛配置值对象
// 仅存储配置信息，不包含计算逻辑
type GameConfig struct {
	GameType    GameType `json:"game_type"`    // 房间类型(台球类型)
	MaxPlayers  int      `json:"max_players"`  // 最大玩家数量
	TargetScore int      `json:"target_score"` // 目标分数（如抢几）
}

var DefaultGameConfig = GameConfig{
	GameType:   GameTypeSnooker,
	MaxPlayers: 2,
}
