package match

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/superwhys/billiard-helper/internal/domain/event"
)

type EightBallStrategy struct{}

type EightBallScoreContext struct {
	StatKey        string `json:"stat_key"`
	ScorerPlayerID uint   `json:"scorer_player_id"`
}

type EightBallScoreEventData event.EventData[EightBallScoreContext]

func NewEightBallStrategy() *EightBallStrategy {
	return &EightBallStrategy{}
}

func (s *EightBallStrategy) defaultPlayerScores() GameScore[map[string]uint] {
	return GameScore[map[string]uint]{
		Score: 0,
		Extra: map[string]uint{},
	}
}

func (s *EightBallStrategy) DefaultScores(ctx context.Context, players []*Player) (json.RawMessage, error) {
	resp := make(map[string]GameScore[map[string]uint])
	for _, player := range players {
		resp[fmt.Sprintf("%d", player.ID)] = s.defaultPlayerScores()
	}

	scoresJSON, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("marshal scores failed: %w", err)
	}

	return scoresJSON, nil
}

func (s *EightBallStrategy) parseCurrentScore(currentScore json.RawMessage) (map[string]GameScore[map[string]uint], error) {
	var currentScoreMap map[string]GameScore[map[string]uint]
	err := json.Unmarshal(currentScore, &currentScoreMap)
	if err != nil {
		return nil, fmt.Errorf("unmarshal current score failed: %w", err)
	}
	return currentScoreMap, nil
}

func (s *EightBallStrategy) CalculateScore(ctx context.Context, players []*Player, currentScore, eventData json.RawMessage) (json.RawMessage, error) {
	var scoreEventData EightBallScoreEventData
	err := json.Unmarshal(eventData, &scoreEventData)
	if err != nil {
		return nil, fmt.Errorf("unmarshal event data failed: %w", err)
	}

	if currentScore == nil {
		currentScore, err = s.DefaultScores(ctx, players)
		if err != nil {
			return nil, fmt.Errorf("default scores failed: %w", err)
		}
	}

	currentScoreMap, err := s.parseCurrentScore(currentScore)
	if err != nil {
		return nil, fmt.Errorf("parse current score failed: %w", err)
	}

	scoreActions := scoreEventData.ScoreActions
	scoreContext := scoreEventData.Context
	if scoreContext.StatKey != "" {
		err := s.increaseTypeCount(currentScoreMap, &scoreContext)
		if err != nil {
			return nil, err
		}
	}

	for _, action := range scoreActions {
		for _, playerID := range action.PlayerIds {
			playerScore, exists := currentScoreMap[fmt.Sprintf("%d", playerID)]
			if !exists {
				playerScore = s.defaultPlayerScores()
			}
			playerScore.Score += action.Score
			currentScoreMap[fmt.Sprintf("%d", playerID)] = playerScore
		}
	}

	return json.Marshal(currentScoreMap)
}

func (s *EightBallStrategy) UndoScore(ctx context.Context, players []*Player, currentScore, eventData json.RawMessage) (json.RawMessage, error) {
	var scoreEventData EightBallScoreEventData
	err := json.Unmarshal(eventData, &scoreEventData)
	if err != nil {
		return nil, fmt.Errorf("unmarshal event data failed: %w", err)
	}

	if currentScore == nil {
		currentScore, err = s.DefaultScores(ctx, players)
		if err != nil {
			return nil, fmt.Errorf("default scores failed: %w", err)
		}
	}

	currentScoreMap, err := s.parseCurrentScore(currentScore)
	if err != nil {
		return nil, fmt.Errorf("parse current score failed: %w", err)
	}

	scoreActions := scoreEventData.ScoreActions
	scoreContext := scoreEventData.Context
	if scoreContext.StatKey != "" {
		err := s.decreaseTypeCount(currentScoreMap, &scoreContext)
		if err != nil {
			return nil, err
		}
	}

	for _, action := range scoreActions {
		for _, playerID := range action.PlayerIds {
			playerScore, exists := currentScoreMap[fmt.Sprintf("%d", playerID)]
			if !exists {
				playerScore = s.defaultPlayerScores()
			}
			playerScore.Score -= action.Score
			currentScoreMap[fmt.Sprintf("%d", playerID)] = playerScore
		}
	}

	return json.Marshal(currentScoreMap)
}

func (s *EightBallStrategy) increaseTypeCount(currentScore map[string]GameScore[map[string]uint], ctx *EightBallScoreContext) error {
	statKey := ctx.StatKey
	if statKey == "" || ctx.ScorerPlayerID == 0 {
		return nil
	}

	playerScore, exists := currentScore[fmt.Sprintf("%d", ctx.ScorerPlayerID)]
	if !exists {
		playerScore = s.defaultPlayerScores()
	}
	if playerScore.Extra == nil {
		playerScore.Extra = map[string]uint{}
	}
	playerScore.Extra[statKey]++
	currentScore[fmt.Sprintf("%d", ctx.ScorerPlayerID)] = playerScore
	return nil
}

func (s *EightBallStrategy) decreaseTypeCount(currentScore map[string]GameScore[map[string]uint], ctx *EightBallScoreContext) error {
	statKey := ctx.StatKey
	if statKey == "" || ctx.ScorerPlayerID == 0 {
		return nil
	}

	playerScore, exists := currentScore[fmt.Sprintf("%d", ctx.ScorerPlayerID)]
	if !exists {
		playerScore = s.defaultPlayerScores()
	}
	if playerScore.Extra == nil {
		playerScore.Extra = map[string]uint{}
	}
	if playerScore.Extra[statKey] > 0 {
		playerScore.Extra[statKey]--
	}
	currentScore[fmt.Sprintf("%d", ctx.ScorerPlayerID)] = playerScore
	return nil
}
