package match

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/miebyte/goutils/logging"
	"github.com/superwhys/billiard-helper/internal/domain/event"
)

// type NineBallScore struct {
// 	Score     int             `json:"score"`
// 	TypeCount map[string]uint `json:"type_count"`
// }

type NineBallScoreContext struct {
	StatKey        string `json:"stat_key"`
	ScorerPlayerID uint   `json:"scorer_player_id"`
}

type NineBallScoreEventData event.EventData[NineBallScoreContext]

type NineBallStrategy struct{}

func NewNineBallStrategy() *NineBallStrategy {
	return &NineBallStrategy{}
}

func (s *NineBallStrategy) defaultPlayerScores() GameScore[map[string]uint] {
	typeCnt := make(map[string]uint)
	for scoreType := range Default9BallScoreConfig() {
		typeCnt[scoreType] = 0
	}

	return GameScore[map[string]uint]{
		Score: 0,
		Extra: typeCnt,
	}
}

func (s *NineBallStrategy) DefaultScores(ctx context.Context, players []*Player) (json.RawMessage, error) {
	resp := make(map[string]GameScore[map[string]uint])

	// 初始化玩家分数
	for _, player := range players {
		// 初始化分数类型计数
		ns := s.defaultPlayerScores()
		resp[fmt.Sprintf("%d", player.ID)] = ns
	}

	scoresJSON, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("marshal scores failed: %w", err)
	}

	return scoresJSON, nil
}

func (s *NineBallStrategy) parseCurrentScore(currentScore json.RawMessage) (map[string]GameScore[map[string]uint], error) {
	var currentScoreMap map[string]GameScore[map[string]uint]
	err := json.Unmarshal(currentScore, &currentScoreMap)
	if err != nil {
		return nil, fmt.Errorf("unmarshal current score failed: %w", err)
	}
	return currentScoreMap, nil
}

func (s *NineBallStrategy) CalculateScore(ctx context.Context, players []*Player, currentScore, eventData json.RawMessage) (json.RawMessage, error) {
	var scoreEventData NineBallScoreEventData
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
	logging.Infoc(ctx, "currentScore: %v", string(currentScore))

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

// UndoScore 回退分数快照
func (s *NineBallStrategy) UndoScore(ctx context.Context, players []*Player, currentScore, eventData json.RawMessage) (json.RawMessage, error) {
	var scoreEventData NineBallScoreEventData
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

func (s *NineBallStrategy) increaseTypeCount(currentScore map[string]GameScore[map[string]uint], ctx *NineBallScoreContext) error {
	statKey := ctx.StatKey
	if statKey == "" {
		return nil
	}

	if ctx.ScorerPlayerID == 0 {
		return nil
	}

	playerScore, exists := currentScore[fmt.Sprintf("%d", ctx.ScorerPlayerID)]
	if !exists {
		playerScore = s.defaultPlayerScores()
	}

	playerScore.Extra[statKey]++
	currentScore[fmt.Sprintf("%d", ctx.ScorerPlayerID)] = playerScore
	return nil
}

func (s *NineBallStrategy) decreaseTypeCount(currentScore map[string]GameScore[map[string]uint], ctx *NineBallScoreContext) error {
	statKey := ctx.StatKey
	if statKey == "" {
		return nil
	}

	if ctx.ScorerPlayerID == 0 {
		return nil
	}

	playerScore, exists := currentScore[fmt.Sprintf("%d", ctx.ScorerPlayerID)]
	if !exists {
		playerScore = s.defaultPlayerScores()
	}

	if playerScore.Extra[statKey] > 0 {
		playerScore.Extra[statKey]--
	}
	currentScore[fmt.Sprintf("%d", ctx.ScorerPlayerID)] = playerScore
	return nil
}
