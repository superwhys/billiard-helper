package match

import (
	"context"
	"encoding/json"
	"testing"
)

type snookerMatchRepo struct {
	IMatchRepository
	updates int
}

func (r *snookerMatchRepo) Update(ctx context.Context, m *Match) error { r.updates++; return nil }

type snookerGameRepo struct {
	IMatchGameRepository
	games []*MatchGame
}

func (r *snookerGameRepo) FindMatchRounds(ctx context.Context, id uint) ([]*MatchGame, error) {
	return r.games, nil
}
func (r *snookerGameRepo) Update(ctx context.Context, g *MatchGame) error { return nil }
func (r *snookerGameRepo) StartGameRound(ctx context.Context, id, round uint, scores json.RawMessage) (*MatchGame, error) {
	g := &MatchGame{ID: round, MatchID: id, GameNum: round, Scores: scores}
	r.games = append(r.games, g)
	return g, nil
}
func frameScores(a, b int) json.RawMessage {
	data, _ := json.Marshal(map[uint]GameScore[map[string]uint]{1: {Score: a}, 2: {Score: b}})
	return data
}
func TestSnookerFramesBestOfAndConcession(t *testing.T) {
	ctx := context.Background()
	mr := &snookerMatchRepo{}
	gr := &snookerGameRepo{games: []*MatchGame{{ID: 1, GameNum: 1, Scores: frameScores(8, 0)}}}
	s := NewMatchService(mr, nil, gr)
	m := &Match{ID: 1, OwnerID: 9, Status: MatchStatusInProgress, MatchType: MatchTypeSnooker, MatchRound: 1, Config: MatchConfig{TargetScore: 3}, Players: []*Player{{ID: 1}, {ID: 2}}}
	if err := s.FinishSnookerFrame(ctx, m, 1, 0); err != nil {
		t.Fatal(err)
	}
	if m.MatchRound != 2 || m.Status != MatchStatusInProgress || *gr.games[0].WinnerID != 1 {
		t.Fatalf("unexpected match %+v", m)
	}
	scores, _ := ParseGameScores(gr.games[1].Scores)
	if scores[1].Score != 0 || scores[2].Score != 0 {
		t.Fatal("next frame not reset")
	}
	if err := s.FinishSnookerFrame(ctx, m, 1, 0); err == nil {
		t.Fatal("duplicate settlement accepted")
	}
	if err := s.FinishSnookerFrame(ctx, m, 2, 0); err == nil {
		t.Fatal("tie accepted")
	}
	// A player may concede while ahead; scores are not rewritten to fake a win.
	gr.games[1].Scores = frameScores(1, 30)
	if err := s.FinishSnookerFrame(ctx, m, 2, 2); err != nil {
		t.Fatal(err)
	}
	if m.Status != MatchStatusFinished || *m.WinnerID != 1 || m.WinnerScore != 2 || len(gr.games) != 2 {
		t.Fatalf("unexpected result %+v", m)
	}
	totals, err := BuildMatchPlayerScores(MatchTypeSnooker, gr.games)
	if err != nil || totals[1] != 2 || totals[2] != 0 {
		t.Fatalf("wrong frame totals: %v %v", totals, err)
	}
	if string(gr.games[1].Scores) != string(frameScores(1, 30)) {
		t.Fatal("concession changed points")
	}
	if err := s.FinishSnookerFrame(ctx, m, 2, 0); err == nil {
		t.Fatal("finished match accepted")
	}
}
func TestSnookerFrameValidation(t *testing.T) {
	for _, target := range []uint{0, 2, 36, 100} {
		m := &Match{MatchType: MatchTypeSnooker, Config: MatchConfig{TargetScore: target}}
		if m.ValidateSnookerConfig() == nil {
			t.Fatalf("accepted target %d", target)
		}
	}
	for _, target := range []uint{1, 3, 5, 35} {
		m := &Match{MatchType: MatchTypeSnooker, Config: MatchConfig{TargetScore: target}}
		if err := m.ValidateSnookerConfig(); err != nil {
			t.Fatal(err)
		}
	}
	gr := &snookerGameRepo{games: []*MatchGame{{GameNum: 1, Scores: frameScores(0, 1)}}}
	s := NewMatchService(&snookerMatchRepo{}, nil, gr)
	m := &Match{Status: MatchStatusInProgress, MatchType: MatchTypeSnooker, MatchRound: 1, Config: MatchConfig{TargetScore: 1}, Players: []*Player{{ID: 1}, {ID: 2}}}
	if err := s.FinishSnookerFrame(context.Background(), m, 1, 99); err == nil {
		t.Fatal("foreign concession accepted")
	}
	if err := s.FinishSnookerFrame(context.Background(), m, 1, 0); err != nil {
		t.Fatal(err)
	}
	if m.Status != MatchStatusFinished || *m.WinnerID != 2 {
		t.Fatal("best of one failed")
	}
}
