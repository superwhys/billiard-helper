package match

type RoomStatus uint8

const (
	RoomStatusPending    RoomStatus = 1 // 未开始
	RoomStatusInProgress RoomStatus = 2 // 进行中
	RoomStatusFinished   RoomStatus = 3 // 已完成
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
	GameType    GameType // 房间类型(台球类型)
	MaxPlayers  int      // 最大玩家数量
	TargetScore int      // 目标分数（如抢几）
}

var DefaultGameConfig = GameConfig{
	GameType:   GameTypeSnooker,
	MaxPlayers: 2,
}
