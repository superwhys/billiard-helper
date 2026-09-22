package match

import (
	"context"
	"time"

	"github.com/superwhys/billiard-helper/internal/errcode"
)

// FinishSnookerFrame settles exactly the requested frame and either starts the
// next one or completes the match. A concession preserves the actual points.
func (s *MatchService) FinishSnookerFrame(ctx context.Context, m *Match, round, concedingPlayerID uint) error {
	if err := m.AssertScoreRequest(m.OwnerID, round); err != nil {
		return err
	}
	if err := m.ValidateSnookerConfig(); err != nil {
		return err
	}
	if len(m.Players) != 2 {
		return errcode.ErrCodeMatchPlayerNotEnough
	}
	games, err := s.matchGameRepository.FindMatchRounds(ctx, m.ID)
	if err != nil {
		return err
	}
	var current *MatchGame
	for _, game := range games {
		if game.GameNum == round {
			current = game
		}
	}
	if current == nil || current.EndAt != 0 {
		return errcode.ErrBadRequest.WithMessage("本局不存在或已经结算")
	}
	scores, err := ParseGameScores(current.Scores)
	if err != nil {
		return err
	}
	// Include a player who has not yet scored in this frame.
	for _, player := range m.Players {
		if _, ok := scores[player.ID]; !ok {
			scores[player.ID] = GameScore[map[string]uint]{}
		}
	}
	winner, _ := FindGameWinner(scores)
	if concedingPlayerID != 0 {
		if concedingPlayerID != m.Players[0].ID && concedingPlayerID != m.Players[1].ID {
			return errcode.ErrBadRequest.WithMessage("认输球员不属于本场比赛")
		}
		for _, player := range m.Players {
			if player.ID != concedingPlayerID {
				id := player.ID
				winner = &id
			}
		}
	}
	if winner == nil {
		return errcode.ErrBadRequest.WithMessage("本局比分相同，请继续记分决出胜负或选择认输")
	}
	current.WinnerID = winner
	current.EndAt = time.Now().Unix()
	wins, err := BuildMatchPlayerScores(MatchTypeSnooker, games)
	if err != nil {
		return err
	}
	if err := s.matchGameRepository.Update(ctx, current); err != nil {
		return err
	}
	m.MatchGames = games
	m.CurrentScores = current.Scores
	if wins[*winner] >= int(m.Config.TargetScore/2+1) {
		m.Status = MatchStatusFinished
		m.WinnerID = winner
		m.WinnerScore = wins[*winner]
	} else {
		if m.IsMaxRoundReached() {
			return errcode.ErrCodeMatchMaxRoundReached
		}
		nextScores, err := (&SnookerStrategy{Config: m.Config}).DefaultScores(ctx, m.Players)
		if err != nil {
			return err
		}
		next, err := s.matchGameRepository.StartGameRound(ctx, m.ID, round+1, nextScores)
		if err != nil {
			return err
		}
		m.MatchGames = append(m.MatchGames, next)
		m.MatchRound++
		m.CurrentScores = nextScores
	}
	return s.matchRepository.Update(ctx, m)
}
