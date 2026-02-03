package match

import "context"

type IMatchService interface {
	// CreateMatch 负责创建比赛的业务流程
	CreateMatch(ctx context.Context, userID uint, config *Match) (*Match, error)
	// FindMatchByID 根据ID查找比赛
	FindMatchByID(ctx context.Context, id uint) (*Match, error)
	// FindPlayerByCode 根据Code查找玩家
	FindPlayerByCode(ctx context.Context, code string) (*Player, error)
	// JoinMatch 处理加入比赛，包括各种校验
	JoinMatch(ctx context.Context, match *Match, player *Player) (*Player, error)
	// StartMatch 开始比赛
	StartMatch(ctx context.Context, match *Match) error
	// EndMatch 结束比赛
	EndMatch(ctx context.Context, match *Match) error
	// LeaveMatch 离开比赛
	LeaveMatch(ctx context.Context, match *Match, player *Player) error
	// KickPlayer 踢出玩家
	KickMatchPlayer(ctx context.Context, match *Match, player *Player) error
}
