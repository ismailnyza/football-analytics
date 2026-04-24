package baseline

import (
	"math"
	"path/filepath"
	"testing"
)

func TestPredictProbabilitiesSumToOne(t *testing.T) {
	probs := predict(1500, 1500, Config{InitialRating: 1500, KFactor: 20, HomeAdvantage: 60, DrawFactor: 0.9, Scale: 400})
	sum := probs.HomeWin + probs.Draw + probs.AwayWin
	if math.Abs(sum-1.0) > 1e-9 {
		t.Fatalf("probabilities sum = %.12f, want 1", sum)
	}
}

func TestRunMatchesDeterministic(t *testing.T) {
	matches := []Match{
		{Season: "1", HomeTeam: "A", AwayTeam: "B", Result: "H"},
		{Season: "1", HomeTeam: "B", AwayTeam: "A", Result: "D"},
	}
	cfg := Config{InitialRating: 1500, KFactor: 20, HomeAdvantage: 60, DrawFactor: 0.9, Scale: 400}
	m1, r1 := RunMatches(matches, cfg, nil)
	m2, r2 := RunMatches(matches, cfg, nil)
	if m1 != m2 {
		t.Fatalf("metrics differ: %#v vs %#v", m1, m2)
	}
	if len(r1) != len(r2) || math.Abs(r1["A"]-r2["A"]) > 1e-9 || math.Abs(r1["B"]-r2["B"]) > 1e-9 {
		t.Fatalf("ratings differ: %#v vs %#v", r1, r2)
	}
}

func TestLoadMatchesFromGlob(t *testing.T) {
	matches, used, err := LoadMatchesFromGlob(filepath.Join("testdata", "*.csv"))
	if err != nil {
		t.Fatalf("LoadMatchesFromGlob error: %v", err)
	}
	if len(used) != 1 {
		t.Fatalf("used files = %d, want 1", len(used))
	}
	if len(matches) != 2 {
		t.Fatalf("matches = %d, want 2", len(matches))
	}
	if matches[0].Season != "sample" {
		t.Fatalf("season = %q, want sample", matches[0].Season)
	}
}
