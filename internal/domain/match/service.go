package match

import (
	"context"
	"fmt"

	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/utils/ptrx"
	"github.com/superwhys/billiard-helper/internal/errcode"
	"github.com/superwhys/billiard-helper/internal/pkg/codegen"
)

type IMatchService interface {
	// CreateMatch 负责创建比赛的业务流程
	CreateMatch(ctx context.Context, userID uint, match *Match) error
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
	// NextRound 下一轮
	NextRound(ctx context.Context, match *Match) error
}

var _ IMatchService = (*MatchService)(nil)

type MatchService struct {
	matchRepository  IMatchRepository
	playerRepository IPlayerRepository
}

func NewMatchService(matchRepository IMatchRepository, playerRepository IPlayerRepository) *MatchService {
	return &MatchService{
		matchRepository:  matchRepository,
		playerRepository: playerRepository,
	}
}

func (s *MatchService) CreateMatch(ctx context.Context, userID uint, match *Match) error {
	if match.IsPlayerOutOfLimit() {
		return errcode.ErrCodeMatchPlayerOutOfLimit
	}

	// 创建比赛
	err := s.matchRepository.Create(ctx, match)
	if err != nil {
		return err
	}

	logging.Infoc(ctx, "create match success, matchID: %d", match.ID)
	// 给玩家赋值比赛ID并生成玩家代码
	for _, player := range match.Players {
		player.MatchID = match.ID

		var payload string
		if player.Type == PlayerTypeReal {
			payload = fmt.Sprintf("%d", ptrx.UintValue(player.UserID))
		} else {
			payload = player.NickName
		}

		player.Code = codegen.GeneratePlayerCode(match.ID, uint8(player.Type), payload)
	}

	// 创建玩家
	err = s.playerRepository.CreateInBatches(ctx, match.Players)
	if err != nil {
		return err
	}

	return nil
}

func (s *MatchService) JoinMatch(ctx context.Context, match *Match, player *Player) (*Player, error) {
	if match.IsPlayerFull() {
		return nil, errcode.ErrCodeMatchPlayerFull
	}

	err := match.JoinPlayer(player)
	if err != nil {
		return nil, err
	}

	err = s.playerRepository.Create(ctx, player)
	if err != nil {
		return nil, err
	}

	return player, nil
}

func (s *MatchService) StartMatch(ctx context.Context, match *Match) error {
	if match.Status != MatchStatusPending {
		return errcode.ErrCodeMatchNotPending
	}

	match.Status = MatchStatusInProgress
	err := s.matchRepository.Update(ctx, match)
	if err != nil {
		return err
	}

	return nil
}

func (s *MatchService) EndMatch(ctx context.Context, match *Match) error {
	if match.Status != MatchStatusInProgress {
		return errcode.ErrCodeMatchNotInProgress
	}

	match.Status = MatchStatusFinished
	err := s.matchRepository.Update(ctx, match)
	if err != nil {
		return err
	}

	return nil
}

func (s *MatchService) LeaveMatch(ctx context.Context, match *Match, player *Player) error {
	if match.Status != MatchStatusPending {
		return errcode.ErrCodeMatchNotPending.WithMessage("比赛已开始，无法退出")
	}

	err := s.playerRepository.Delete(ctx, player.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *MatchService) KickMatchPlayer(ctx context.Context, match *Match, player *Player) error {
	return s.LeaveMatch(ctx, match, player)
}

func (s *MatchService) NextRound(ctx context.Context, match *Match) error {
	if match.Status != MatchStatusInProgress {
		return errcode.ErrCodeMatchNotInProgress
	}

	maxRound := match.Config.TargetScore
	if match.MatchRound >= maxRound {
		return errcode.ErrCodeMatchMaxRoundReached
	}

	match.MatchRound++
	err := s.matchRepository.Update(ctx, match)
	if err != nil {
		return err
	}

	return nil
}
