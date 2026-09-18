package match

import (
	"context"
	"encoding/json"
)

type MatchTypeStrategy interface {
	DefaultScores(ctx context.Context, players []*Player) (json.RawMessage, error)
	CalculateScore(ctx context.Context, players []*Player, currentScore, eventData json.RawMessage) (json.RawMessage, error)
	// UndoScore 回退分数快照
	UndoScore(ctx context.Context, players []*Player, currentScore, eventData json.RawMessage) (json.RawMessage, error)
}

func MatchTypeStrategyFactory(matchType MatchType) MatchTypeStrategy {
	switch matchType {
	case MatchTypeSnooker:
		return &SnookerStrategy{}
	case MatchType8Ball:
		return NewEightBallStrategy()
	case MatchType9Ball:
		return NewNineBallStrategy()
	default:
		return nil
	}
}
