package match

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/superwhys/billiard-helper/internal/domain/event"
)

func snookerEvent(key string, scorer uint, recipient uint, points int) json.RawMessage {
	data, _ := json.Marshal(event.EventData[map[string]any]{
		ScoreActions: []event.ScoreAction{{PlayerIds: []uint{recipient}, Score: points}},
		Context:      map[string]any{"stat_key": key, "scorer_player_id": scorer},
	})
	return data
}

func TestSnookerScoringAndUndo(t *testing.T) {
	strategy := MatchTypeStrategyFactory(MatchTypeSnooker)
	if strategy == nil {
		t.Fatal("snooker strategy is not registered")
	}
	ctx := context.Background()
	players := []*Player{{ID: 1}, {ID: 2}}
	initial, err := strategy.DefaultScores(ctx, players)
	if err != nil {
		t.Fatal(err)
	}
	for key, points := range map[string]int{"red": 1, "yellow": 2, "green": 3, "brown": 4, "blue": 5, "pink": 6, "black": 7, "foul": 4} {
		t.Run(key, func(t *testing.T) {
			recipient := uint(1)
			if key == "foul" {
				recipient = 2
			}
			data := snookerEvent(key, 1, recipient, points)
			scored, err := strategy.CalculateScore(ctx, players, initial, data)
			if err != nil {
				t.Fatal(err)
			}
			scores, _ := ParseGameScores(scored)
			if scores[recipient].Score != points || scores[1].Extra[key] != 1 {
				t.Fatalf("unexpected scores: %s", scored)
			}
			undone, err := strategy.UndoScore(ctx, players, scored, data)
			if err != nil {
				t.Fatal(err)
			}
			if string(undone) != string(initial) {
				t.Fatalf("undo mismatch: %s != %s", undone, initial)
			}
		})
	}
	for _, points := range []int{5, 6, 7} {
		if _, err := strategy.CalculateScore(ctx, players, initial, snookerEvent("foul", 1, 2, points)); err != nil {
			t.Fatal(err)
		}
	}
	// A player added after room creation must still receive an initialized score.
	partial, _ := strategy.DefaultScores(ctx, players[:1])
	scored, err := strategy.CalculateScore(ctx, players, partial, snookerEvent("red", 2, 2, 1))
	if err != nil {
		t.Fatal(err)
	}
	scores, _ := ParseGameScores(scored)
	if len(scores) != 2 || scores[2].Score != 1 {
		t.Fatal(string(scored))
	}
}

func TestSnookerRejectsInvalidScores(t *testing.T) {
	s := MatchTypeStrategyFactory(MatchTypeSnooker)
	if s == nil {
		t.Fatal("snooker strategy is not registered")
	}
	ctx := context.Background()
	players := []*Player{{ID: 1}, {ID: 2}}
	initial, _ := s.DefaultScores(ctx, players)
	invalid := []json.RawMessage{
		snookerEvent("black", 1, 1, 8), snookerEvent("red", 1, 2, 1),
		snookerEvent("red", 99, 1, 1), snookerEvent("red", 1, 99, 1),
		snookerEvent("red", 1, 1, -1), snookerEvent("red", 1, 1, 0),
		snookerEvent("foul", 1, 1, 4), snookerEvent("foul", 1, 2, 3),
		snookerEvent("foul", 1, 2, 8), snookerEvent("unknown", 1, 1, 1),
		json.RawMessage(`{}`), json.RawMessage(`null`), json.RawMessage(`{"score_actions":[{"player_ids":[1,1],"score":1}],"context":{"stat_key":"red","scorer_player_id":1}}`),
		json.RawMessage(`{"score_actions":[{"player_ids":[1],"score":1},{"player_ids":[2],"score":1}],"context":{"stat_key":"red","scorer_player_id":1}}`),
	}
	for _, data := range invalid {
		if _, err := s.CalculateScore(ctx, players, initial, data); err == nil {
			t.Errorf("accepted %s", data)
		}
	}
	if _, err := s.UndoScore(ctx, players, initial, snookerEvent("red", 1, 1, 1)); err == nil {
		t.Fatal("undo underflow accepted")
	}
	if _, err := s.CalculateScore(ctx, players[:1], initial, snookerEvent("red", 1, 1, 1)); err == nil {
		t.Fatal("accepted one player")
	}
}
