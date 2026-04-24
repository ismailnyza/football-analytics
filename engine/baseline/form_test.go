package baseline

import (
	"math"
	"testing"
)

func TestExtractTeamForm(t *testing.T) {
	matches := []Match{
		{HomeTeam: "A", AwayTeam: "B", HomeGoals: 2, AwayGoals: 0, Result: "H"},
		{HomeTeam: "A", AwayTeam: "C", HomeGoals: 1, AwayGoals: 1, Result: "D"},
		{HomeTeam: "D", AwayTeam: "A", HomeGoals: 3, AwayGoals: 1, Result: "H"},
		{HomeTeam: "A", AwayTeam: "E", HomeGoals: 2, AwayGoals: 1, Result: "H"},
	}
	form := ExtractTeamForm(matches, 4)
	a, ok := form["A"]
	if !ok {
		t.Fatalf("team A missing from form")
	}
	if a.RecentWins != 2 || a.RecentDraws != 1 || a.RecentLosses != 1 {
		t.Fatalf("A record = %dW %dD %dL, want 2W 1D 1L", a.RecentWins, a.RecentDraws, a.RecentLosses)
	}
	if a.RecentGF != 6 || a.RecentGA != 5 {
		t.Fatalf("A goals = %dGF %dGA, want 6GF 5GA", a.RecentGF, a.RecentGA)
	}
	if math.Abs(a.PointsPerGame-1.75) > 1e-9 {
		t.Fatalf("A PPG = %.4f, want 1.75", a.PointsPerGame)
	}
}

func TestExtractTeamFormWindowShorterThanHistory(t *testing.T) {
	matches := []Match{
		{HomeTeam: "A", AwayTeam: "B", HomeGoals: 1, AwayGoals: 0, Result: "H"},
		{HomeTeam: "C", AwayTeam: "A", HomeGoals: 2, AwayGoals: 0, Result: "H"},
		{HomeTeam: "A", AwayTeam: "D", HomeGoals: 3, AwayGoals: 1, Result: "H"},
	}
	form := ExtractTeamForm(matches, 2)
	a, ok := form["A"]
	if !ok {
		t.Fatalf("team A missing")
	}
	if a.Window != 2 {
		t.Fatalf("window = %d, want 2", a.Window)
	}
	if a.RecentWins != 1 || a.RecentLosses != 1 {
		t.Fatalf("recent 2 = %dW %dL, want 1W 1L", a.RecentWins, a.RecentLosses)
	}
}

func TestCalculateFormDelta(t *testing.T) {
	home := TeamFormRecord{PointsPerGame: 2.5, GoalDiffPerGame: 1.2}
	away := TeamFormRecord{PointsPerGame: 1.0, GoalDiffPerGame: -0.5}
	delta := CalculateFormDelta(home, away)
	expected := (2.5-1.0)*0.5 + (1.2-(-0.5))*0.3
	if math.Abs(delta-expected) > 1e-9 {
		t.Fatalf("form delta = %.4f, want %.4f", delta, expected)
	}
}

func TestFormExtractionDeterministic(t *testing.T) {
	matches := []Match{
		{HomeTeam: "A", AwayTeam: "B", HomeGoals: 1, AwayGoals: 0, Result: "H"},
		{HomeTeam: "B", AwayTeam: "A", HomeGoals: 2, AwayGoals: 1, Result: "H"},
	}
	f1 := ExtractTeamForm(matches, 2)
	f2 := ExtractTeamForm(matches, 2)
	if f1["A"].PointsPerGame != f2["A"].PointsPerGame {
		t.Fatalf("non-deterministic form")
	}
}
