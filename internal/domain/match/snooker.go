package match

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/superwhys/billiard-helper/internal/domain/event"
	"github.com/superwhys/billiard-helper/internal/errcode"
)

// SnookerStrategy records points awarded by the scorer; table state and refereeing
// (including free balls and misses) remain decisions made at the table.
type SnookerStrategy struct{}

type SnookerScoreContext struct {
	StatKey        string `json:"stat_key"`
	ScorerPlayerID uint   `json:"scorer_player_id"` // Player potting or committing the foul.
}

var snookerBallPoints = map[string]int{
	"red": 1, "yellow": 2, "green": 3, "brown": 4,
	"blue": 5, "pink": 6, "black": 7,
}

func (s *SnookerStrategy) DefaultScores(ctx context.Context, players []*Player) (json.RawMessage, error) {
	scores := make(map[string]GameScore[map[string]uint], len(players))
	for _, player := range players {
		scores[strconv.FormatUint(uint64(player.ID), 10)] = GameScore[map[string]uint]{Extra: map[string]uint{}}
	}
	return json.Marshal(scores)
}

func (s *SnookerStrategy) CalculateScore(ctx context.Context, players []*Player, currentScore, eventData json.RawMessage) (json.RawMessage, error) {
	return s.apply(players, currentScore, eventData, false)
}

func (s *SnookerStrategy) UndoScore(ctx context.Context, players []*Player, currentScore, eventData json.RawMessage) (json.RawMessage, error) {
	return s.apply(players, currentScore, eventData, true)
}

func (s *SnookerStrategy) apply(players []*Player, currentScore, eventData json.RawMessage, undo bool) (json.RawMessage, error) {
	invalid := errcode.ErrBadRequest.WithMessage("斯诺克计分无效，请检查球员、球色和分值")
	if len(players) != 2 || players[0].ID == players[1].ID {
		return nil, invalid
	}
	var data event.EventData[SnookerScoreContext]
	if err := json.Unmarshal(eventData, &data); err != nil {
		return nil, invalid
	}
	if len(data.ScoreActions) != 1 || len(data.ScoreActions[0].PlayerIds) != 1 {
		return nil, invalid
	}
	scorer := data.Context.ScorerPlayerID
	recipient := scorer
	if scorer != players[0].ID && scorer != players[1].ID {
		return nil, invalid
	}
	action := data.ScoreActions[0]
	key := data.Context.StatKey
	if key == "foul" {
		if action.Score < 4 || action.Score > 7 {
			return nil, invalid
		}
		for _, player := range players {
			if player.ID != scorer {
				recipient = player.ID
			}
		}
	} else if points, ok := snookerBallPoints[key]; !ok || action.Score != points {
		return nil, invalid
	}
	if action.PlayerIds[0] != recipient {
		return nil, invalid
	}

	scores, err := ParseGameScores(currentScore)
	if err != nil {
		return nil, fmt.Errorf("parse snooker scores: %w", err)
	}
	// The initial snapshot is created before additional players join the room.
	for _, player := range players {
		value := scores[player.ID]
		if value.Extra == nil {
			value.Extra = map[string]uint{}
		}
		scores[player.ID] = value
	}
	source := scores[scorer]
	target := scores[recipient]
	if undo {
		if source.Extra[key] == 0 || target.Score < action.Score {
			return nil, errcode.ErrBadRequest.WithMessage("无法撤销不匹配的斯诺克计分")
		}
		source.Extra[key]--
		if source.Extra[key] == 0 {
			delete(source.Extra, key)
		}
		target.Score -= action.Score
	} else {
		source.Extra[key]++
		target.Score += action.Score
	}
	scores[scorer] = source
	// Preserve the stat update when scorer and recipient are the same player.
	target.Extra = scores[recipient].Extra
	scores[recipient] = target
	return json.Marshal(scores)
}

func (m *Match) ValidateSnookerConfig() error {
	if m.MatchType != MatchTypeSnooker {
		return nil
	}
	if m.Config.TargetScore == 0 || m.Config.TargetScore > 35 || m.Config.TargetScore%2 == 0 {
		return errcode.ErrBadRequest.WithMessage("斯诺克赛制必须为 1 到 35 的奇数局")
	}
	return nil
}
