package match

import (
	"encoding/json"

	"github.com/superwhys/billiard-helper/internal/domain/event"
	"github.com/superwhys/billiard-helper/internal/errcode"
)

type SnookerState struct {
	RedCount        int    `json:"red_count"`
	RedsRemaining   int    `json:"reds_remaining"`
	NextBall        string `json:"next_ball"`
	ActivePlayerID  uint   `json:"active_player_id"`
	BreakScore      int    `json:"break_score"`
	RemainingPoints int    `json:"remaining_points"`
	CanUndo         bool   `json:"can_undo"`
}

func (c MatchConfig) SnookerRedCount() (int, error) {
	value, exists := c.Data["red_count"]
	if !exists {
		return 15, nil
	}
	raw, err := json.Marshal(value)
	var count int
	if err != nil || json.Unmarshal(raw, &count) != nil || count < 1 || count > 15 {
		return 0, errcode.ErrBadRequest.WithMessage("红球数量必须为 1 到 15 的整数")
	}
	return count, nil
}

func ReadSnookerState(snapshot json.RawMessage) (*SnookerState, error) {
	if len(snapshot) == 0 {
		return nil, nil
	}
	var envelope struct {
		State *SnookerState `json:"_snooker"`
	}
	if err := json.Unmarshal(snapshot, &envelope); err != nil {
		return nil, err
	}
	return envelope.State, nil
}

func marshalSnooker(scores map[uint]GameScore[map[string]uint], state *SnookerState) (json.RawMessage, error) {
	raw, err := json.Marshal(scores)
	if err != nil {
		return nil, err
	}
	var snapshot map[string]json.RawMessage
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		return nil, err
	}
	snapshot["_snooker"], err = json.Marshal(state)
	if err != nil {
		return nil, err
	}
	return json.Marshal(snapshot)
}

func (s *SnookerState) updateRemaining() {
	switch s.NextBall {
	case "red":
		s.RemainingPoints = s.RedsRemaining*8 + 27
	case "colour":
		s.RemainingPoints = s.RedsRemaining*8 + 34
	case "done":
		s.RemainingPoints = 0
	case "respotted_black":
		s.RemainingPoints = 7
	default:
		s.RemainingPoints = 0
		for value := snookerBallPoints[s.NextBall]; value <= 7; value++ {
			s.RemainingPoints += value
		}
	}
}

func (s *SnookerState) endBreak(next uint) {
	s.ActivePlayerID = next
	s.BreakScore = 0
	if s.NextBall == "colour" {
		s.NextBall = "red"
		if s.RedsRemaining == 0 {
			s.NextBall = "yellow"
		}
	}
}

func (s *SnookerStrategy) calculate(players []*Player, snapshot, eventData json.RawMessage) (json.RawMessage, error) {
	state, err := ReadSnookerState(snapshot)
	if err != nil {
		return nil, err
	}
	// Existing frames have no shot-order information. Preserve their manual
	// scoring until settlement; the next frame starts with the new rules.
	if state == nil {
		return s.apply(players, snapshot, eventData, false)
	}
	invalid := errcode.ErrBadRequest.WithMessage("斯诺克计分无效，请检查球员、球色和分值")
	if len(players) != 2 || players[0].ID == players[1].ID {
		return nil, invalid
	}
	var data event.EventData[SnookerScoreContext]
	if err := json.Unmarshal(eventData, &data); err != nil {
		return nil, invalid
	}
	scorer := data.Context.ScorerPlayerID
	if scorer != players[0].ID && scorer != players[1].ID {
		return nil, invalid
	}
	if state.ActivePlayerID != 0 && state.ActivePlayerID != scorer {
		return nil, errcode.ErrBadRequest.WithMessage("击球球员已变化，请刷新后重试")
	}
	if state.NextBall == "done" {
		return nil, errcode.ErrBadRequest.WithMessage("台面已清空，请结算本局或撤销上一笔")
	}
	opponent := players[0].ID
	if opponent == scorer {
		opponent = players[1].ID
	}
	key := data.Context.StatKey
	scores, err := ParseGameScores(snapshot)
	if err != nil {
		return nil, err
	}
	for _, player := range players {
		score := scores[player.ID]
		if score.Extra == nil {
			score.Extra = map[string]uint{}
		}
		scores[player.ID] = score
	}
	if key == "turn_end" {
		next := data.Context.NextPlayerID
		if len(data.ScoreActions) != 0 || data.Context.RedsRemoved != 0 || next != opponent {
			return nil, invalid
		}
		state.endBreak(next)
	} else {
		if len(data.ScoreActions) != 1 || len(data.ScoreActions[0].PlayerIds) != 1 {
			return nil, invalid
		}
		action := data.ScoreActions[0]
		recipient := scorer
		if key == "foul" {
			recipient = opponent
		}
		if action.PlayerIds[0] != recipient {
			return nil, invalid
		}
		if key == "foul" {
			minimum := 4
			if points := snookerBallPoints[state.NextBall]; points > minimum {
				minimum = points
			}
			if state.NextBall == "respotted_black" {
				minimum = 7
			}
			if action.Score < minimum || action.Score > 7 || data.Context.RedsRemoved < 0 || data.Context.RedsRemoved > state.RedsRemaining {
				return nil, invalid
			}
			next := data.Context.NextPlayerID
			if next == 0 {
				next = opponent
			}
			if next != scorer && next != opponent {
				return nil, invalid
			}
			state.RedsRemaining -= data.Context.RedsRemoved
			state.endBreak(next)
			if state.RedsRemaining == 0 && state.NextBall == "red" {
				state.NextBall = "yellow"
			}
		} else {
			points, ok := snookerBallPoints[key]
			if !ok || action.Score != points || data.Context.RedsRemoved != 0 || data.Context.NextPlayerID != 0 {
				return nil, invalid
			}
			expected := state.NextBall
			valid := key == expected || expected == "colour" && key != "red" || expected == "respotted_black" && key == "black"
			if !valid {
				return nil, errcode.ErrBadRequest.WithMessage("进球顺序不正确，请按当前目标球记分")
			}
			state.ActivePlayerID = scorer
			state.BreakScore += points
			switch expected {
			case "red":
				if state.RedsRemaining <= 0 {
					return nil, invalid
				}
				state.RedsRemaining--
				state.NextBall = "colour"
			case "colour":
				state.NextBall = "red"
				if state.RedsRemaining == 0 {
					state.NextBall = "yellow"
				}
			default:
				state.NextBall = "done"
				for colour, value := range snookerBallPoints {
					if value == points+1 {
						state.NextBall = colour
					}
				}
			}
		}
		source := scores[scorer]
		source.Extra[key]++
		if uint(state.BreakScore) > source.Extra["highest_break"] {
			source.Extra["highest_break"] = uint(state.BreakScore)
		}
		scores[scorer] = source
		target := scores[recipient]
		target.Score += action.Score
		scores[recipient] = target
		// Potting the final black or fouling with only black left ends the frame;
		// tied points require a re-spotted black, with striker chosen at the table.
		if state.NextBall == "done" || key == "foul" && (state.NextBall == "black" || state.NextBall == "respotted_black") {
			state.NextBall = "done"
			if scores[scorer].Score == scores[opponent].Score {
				state.NextBall = "respotted_black"
				state.BreakScore = 0
				state.ActivePlayerID = 0
			}
		}
	}
	state.CanUndo = true
	state.updateRemaining()
	return marshalSnooker(scores, state)
}
