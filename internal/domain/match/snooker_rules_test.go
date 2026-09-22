package match

import (
	"context"
	"encoding/json"
	"testing"
)

func TestSnookerRedColourOrder(t *testing.T) {
	strategy := MatchTypeStrategyFactory(MatchTypeSnooker)
	players := []*Player{{ID: 1}, {ID: 2}}
	scores, _ := strategy.DefaultScores(context.Background(), players)
	if _, err := strategy.CalculateScore(context.Background(), players, scores, snookerEvent("black", 1, 1, 7)); err == nil {
		t.Fatal("a frame must start with a red")
	}
	scores, err := strategy.CalculateScore(context.Background(), players, scores, snookerEvent("red", 1, 1, 1))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := strategy.CalculateScore(context.Background(), players, scores, snookerEvent("red", 1, 1, 1)); err == nil {
		t.Fatal("a red must be followed by a colour")
	}
	var snapshot map[string]json.RawMessage
	json.Unmarshal(scores, &snapshot)
	if len(snapshot["_snooker"]) == 0 {
		t.Fatal("missing persistent table state")
	}
}

func TestSnookerClearanceAndUndo(t *testing.T) {
	ctx := context.Background()
	players := []*Player{{ID: 1}, {ID: 2}}
	s := &SnookerStrategy{Config: MatchConfig{Data: map[string]any{"red_count": 1}}}
	scores, err := s.DefaultScores(ctx, players)
	if err != nil {
		t.Fatal(err)
	}
	assertState := func(reds, remaining, single int, next string) {
		t.Helper()
		st, err := ReadSnookerState(scores)
		if err != nil || st == nil {
			t.Fatalf("state %s: %v", scores, err)
		}
		if st.RedsRemaining != reds || st.RemainingPoints != remaining || st.BreakScore != single || st.NextBall != next {
			t.Fatalf("unexpected state %+v", st)
		}
	}
	assertState(1, 35, 0, "red")
	var history []json.RawMessage
	var events []json.RawMessage
	apply := func(key string, points int) {
		t.Helper()
		data := snookerEvent(key, 1, 1, points)
		var parsed map[string]json.RawMessage
		json.Unmarshal(data, &parsed)
		parsed["before_scores"] = scores
		data, _ = json.Marshal(parsed)
		history = append(history, scores)
		events = append(events, data)
		scores, err = s.CalculateScore(ctx, players, scores, data)
		if err != nil {
			t.Fatal(err)
		}
	}
	apply("red", 1)
	assertState(0, 34, 1, "colour")
	apply("black", 7)
	assertState(0, 27, 8, "yellow")
	if _, err := s.CalculateScore(ctx, players, scores, snookerEvent("pink", 1, 1, 6)); err == nil {
		t.Fatal("out-of-order clearance accepted")
	}
	for _, key := range []string{"yellow", "green", "brown", "blue", "pink", "black"} {
		apply(key, snookerBallPoints[key])
	}
	assertState(0, 0, 35, "done")
	if _, err := s.CalculateScore(ctx, players, scores, snookerEvent("black", 1, 1, 7)); err == nil {
		t.Fatal("scoring after clearance accepted")
	}
	// Restore every prior stage, score and break exactly, including the last red.
	for i := len(events) - 1; i >= 0; i-- {
		scores, err = s.UndoScore(ctx, players, scores, events[i])
		if err != nil {
			t.Fatal(err)
		}
		if string(scores) != string(history[i]) {
			t.Fatal("undo did not restore entire snapshot")
		}
	}
	assertState(1, 35, 0, "red")
}

func TestSnookerTurnAndFoul(t *testing.T) {
	ctx := context.Background()
	players := []*Player{{ID: 1}, {ID: 2}}
	s := &SnookerStrategy{Config: MatchConfig{Data: map[string]any{"red_count": 6}}}
	scores, _ := s.DefaultScores(ctx, players)
	st, _ := ReadSnookerState(scores)
	if st.RemainingPoints != 75 {
		t.Fatal(st)
	}
	scores, _ = s.CalculateScore(ctx, players, scores, snookerEvent("red", 1, 1, 1))
	if _, err := s.CalculateScore(ctx, players, scores, snookerEvent("black", 2, 2, 7)); err == nil {
		t.Fatal("another player bypassed turn")
	}
	turn := json.RawMessage(`{"score_actions":[],"context":{"stat_key":"turn_end","scorer_player_id":1,"next_player_id":2}}`)
	scores, err := s.CalculateScore(ctx, players, scores, turn)
	if err != nil {
		t.Fatal(err)
	}
	st, _ = ReadSnookerState(scores)
	if st.NextBall != "red" || st.ActivePlayerID != 2 || st.BreakScore != 0 || st.RemainingPoints != 67 {
		t.Fatal(st)
	}
	scores, _ = s.CalculateScore(ctx, players, scores, snookerEvent("red", 2, 2, 1))
	// Foul: opponent receives points, one red stays off the table, striker asked to play again.
	foul := json.RawMessage(`{"score_actions":[{"player_ids":[1],"score":4}],"context":{"stat_key":"foul","scorer_player_id":2,"reds_removed":1,"next_player_id":2}}`)
	scores, err = s.CalculateScore(ctx, players, scores, foul)
	if err != nil {
		t.Fatal(err)
	}
	st, _ = ReadSnookerState(scores)
	values, _ := ParseGameScores(scores)
	if st.BreakScore != 0 || st.RedsRemaining != 3 || st.NextBall != "red" || st.RemainingPoints != 51 || st.ActivePlayerID != 2 || values[1].Score != 5 {
		t.Fatalf("%s", scores)
	}
}

func TestSnookerLastRedMissAndRespottedBlack(t *testing.T) {
	ctx := context.Background()
	players := []*Player{{ID: 1}, {ID: 2}}
	s := &SnookerStrategy{Config: MatchConfig{Data: map[string]any{"red_count": 1}}}
	scores, _ := s.DefaultScores(ctx, players)
	scores, _ = s.CalculateScore(ctx, players, scores, snookerEvent("red", 1, 1, 1))
	turn := json.RawMessage(`{"context":{"stat_key":"turn_end","scorer_player_id":1,"next_player_id":2}}`)
	scores, err := s.CalculateScore(ctx, players, scores, turn)
	if err != nil {
		t.Fatal(err)
	}
	st, _ := ReadSnookerState(scores)
	if st.NextBall != "yellow" || st.RemainingPoints != 27 {
		t.Fatal(st)
	}
	// A foul that removes the final red also moves directly to yellow.
	scores, _ = s.DefaultScores(ctx, players)
	scores, err = s.CalculateScore(ctx, players, scores, json.RawMessage(`{"score_actions":[{"player_ids":[2],"score":4}],"context":{"stat_key":"foul","scorer_player_id":1,"reds_removed":1}}`))
	if err != nil {
		t.Fatal(err)
	}
	st, _ = ReadSnookerState(scores)
	if st.NextBall != "yellow" || st.RedsRemaining != 0 {
		t.Fatal(st)
	}
	scores, _ = marshalSnooker(map[uint]GameScore[map[string]uint]{1: {Score: 0}, 2: {Score: 7}}, &SnookerState{RedCount: 1, NextBall: "black", RemainingPoints: 7})
	if _, err := s.CalculateScore(ctx, players, scores, snookerEvent("foul", 1, 2, 4)); err == nil {
		t.Fatal("black foul under 7 accepted")
	}
	scores, err = s.CalculateScore(ctx, players, scores, snookerEvent("black", 1, 1, 7))
	if err != nil {
		t.Fatal(err)
	}
	st, _ = ReadSnookerState(scores)
	if st.NextBall != "respotted_black" || st.RemainingPoints != 7 || st.BreakScore != 0 {
		t.Fatal(st)
	}
	scores, err = s.CalculateScore(ctx, players, scores, snookerEvent("black", 2, 2, 7))
	if err != nil {
		t.Fatal(err)
	}
	st, _ = ReadSnookerState(scores)
	if st.NextBall != "done" {
		t.Fatal(st)
	}
}

func TestSnookerRedCountValidation(t *testing.T) {
	for _, value := range []any{0, 16, -1, 2.5, "6", nil} {
		if _, err := (MatchConfig{Data: map[string]any{"red_count": value}}).SnookerRedCount(); err == nil {
			t.Fatalf("accepted %#v", value)
		}
	}
	for _, value := range []any{1, 6, 10, 15, float64(6), json.Number("10")} {
		if _, err := (MatchConfig{Data: map[string]any{"red_count": value}}).SnookerRedCount(); err != nil {
			t.Fatal(err)
		}
	}
}
