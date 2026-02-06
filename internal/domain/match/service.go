package match

import (
	"context"

	"github.com/superwhys/billiard-helper/internal/errcode"
)

type IMatchService interface {
	// CreateMatch 负责创建比赛的业务流程
	CreateMatch(ctx context.Context, userID uint, match *Match) (*Match, error)
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

func (s *MatchService) CreateMatch(ctx context.Context, userID uint, match *Match) (*Match, error) {
	if match.IsPlayerOutOfLimit() {
		return nil, errcode.ErrCodeMatchPlayerOutOfLimit
	}

	err := s.matchRepository.Create(ctx, match)
	if err != nil {
		return nil, err
	}

	return match, nil
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
