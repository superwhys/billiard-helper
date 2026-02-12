package match

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/miebyte/goutils/logging"
	"github.com/miebyte/goutils/utils/ptrx"
	"github.com/superwhys/billiard-helper/internal/domain/event"
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
	// NextRound 下一轮
	NextRound(ctx context.Context, matchID uint) (*Match, error)
	// CalculateMatchGameScore 计算比赛轮次中的分数快照
	// 计算分数快照的逻辑：
	// 1. 获取当前的分数快照
	// 2. 遍历当前事件的 ScoreActions，根据 ScoreActions 中的 PlayerIds 和 Score 计算新的分数快照
	//     2.1 这里也许要使用策略模式，不同的玩法有不同的分数计算逻辑
	//     2.2 然后这里传递当前房间的玩家，当前的分数快照，以及分数事件数据，来计算新的分数快照
	CalculateMatchGameScore(ctx context.Context, match *Match, matchGame *MatchGame, event *event.Event) (json.RawMessage, error)
	UndoMatchGameScore(ctx context.Context, match *Match, matchGame *MatchGame, event *event.Event) (json.RawMessage, error)
}

var _ IMatchService = (*MatchService)(nil)

type MatchService struct {
	matchRepository     IMatchRepository
	matchGameRepository IMatchGameRepository
	playerRepository    IPlayerRepository
}

func NewMatchService(
	matchRepository IMatchRepository,
	playerRepository IPlayerRepository,
	matchGameRepository IMatchGameRepository,
) *MatchService {
	return &MatchService{
		matchRepository:     matchRepository,
		playerRepository:    playerRepository,
		matchGameRepository: matchGameRepository,
	}
}

func (s *MatchService) CreateMatch(ctx context.Context, userID uint, match *Match) error {
	if match.IsPlayerOutOfLimit() {
		return errcode.ErrCodeMatchPlayerOutOfLimit
	}

	if match.MatchType == MatchType9Ball {
		match.Config.Data = Default9BallScoreConfig()
	}

	// 创建比赛
	err := s.matchRepository.Create(ctx, match)
	if err != nil {
		return fmt.Errorf("create match failed: %w", err)
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
		return fmt.Errorf("create player failed: %w", err)
	}

	// 根据不同玩法初始化默认分数快照
	strategy := MatchTypeStrategyFactory(match.MatchType)
	defaultScores, err := strategy.DefaultScores(ctx, match.Players)
	if err != nil {
		return fmt.Errorf("default scores failed: %w", err)
	}

	_, err = s.matchGameRepository.StartGameRound(ctx, match.ID, 1, defaultScores)
	if err != nil {
		return fmt.Errorf("start game round failed: %w", err)
	}

	return nil
}

func (s *MatchService) JoinMatch(ctx context.Context, match *Match, player *Player) (*Player, error) {
	if match.IsPlayerFull() {
		return nil, errcode.ErrCodeMatchPlayerFull
	}

	err := match.JoinPlayer(player)
	if err != nil {
		return nil, fmt.Errorf("join match failed: %w", err)
	}

	err = s.playerRepository.Create(ctx, player)
	if err != nil {
		return nil, fmt.Errorf("create player failed: %w", err)
	}

	return player, nil
}

func (s *MatchService) StartMatch(ctx context.Context, match *Match) error {
	if err := match.AssertStartable(); err != nil {
		return err
	}

	match.Status = MatchStatusInProgress
	err := s.matchRepository.Update(ctx, match)
	if err != nil {
		return fmt.Errorf("start match failed: %w", err)
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
		return fmt.Errorf("end match failed: %w", err)
	}

	return nil
}

func (s *MatchService) LeaveMatch(ctx context.Context, match *Match, player *Player) error {
	if match.Status != MatchStatusPending {
		return errcode.ErrCodeMatchNotPending.WithMessage("比赛已开始，无法退出")
	}

	err := s.playerRepository.Delete(ctx, player.ID)
	if err != nil {
		return fmt.Errorf("leave match failed: %w", err)
	}

	return nil
}

func (s *MatchService) NextRound(ctx context.Context, matchID uint) (*Match, error) {
	match, err := s.matchRepository.FindByID(ctx, matchID, false)
	if err != nil {
		return nil, fmt.Errorf("find match failed: %w", err)
	}

	if !match.IsStart() {
		return nil, errcode.ErrCodeMatchNotInProgress
	}

	if match.IsMaxRoundReached() {
		return nil, errcode.ErrCodeMatchMaxRoundReached
	}

	beforeRound := match.MatchRound
	match.MatchRound++
	err = s.matchRepository.Update(ctx, match)
	if err != nil {
		return nil, fmt.Errorf("update match failed: %w", err)
	}

	// TODO: 这里应该是要获取到本轮赢的玩家的
	// 结束上一轮
	err = s.matchGameRepository.EndGameRound(ctx, match.ID, beforeRound)
	if err != nil {
		return nil, fmt.Errorf("end game round failed: %w", err)
	}

	// 开始新的一轮
	strategy := MatchTypeStrategyFactory(match.MatchType)
	defaultScores, err := strategy.DefaultScores(ctx, match.Players)
	if err != nil {
		return nil, fmt.Errorf("default scores failed: %w", err)
	}
	_, err = s.matchGameRepository.StartGameRound(ctx, match.ID, match.MatchRound, defaultScores)
	if err != nil {
		return nil, fmt.Errorf("start game round failed: %w", err)
	}

	return match, nil
}

func (s *MatchService) CalculateMatchGameScore(ctx context.Context, match *Match, matchGame *MatchGame, event *event.Event) (json.RawMessage, error) {
	strategy := MatchTypeStrategyFactory(match.MatchType)

	newScores, err := strategy.CalculateScore(ctx, match.Players, matchGame.Scores, event.Data)
	if err != nil {
		return nil, err
	}

	return newScores, nil
}

func (s *MatchService) UndoMatchGameScore(ctx context.Context, match *Match, matchGame *MatchGame, event *event.Event) (json.RawMessage, error) {
	strategy := MatchTypeStrategyFactory(match.MatchType)

	newScores, err := strategy.UndoScore(ctx, match.Players, matchGame.Scores, event.Data)
	if err != nil {
		return nil, err
	}

	return newScores, nil
}
