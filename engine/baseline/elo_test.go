package baseline

import (
	"math"
	"path/filepath"
	"testing"
)

func TestPredictProbabilitiesSumToOne(t *testing.T) {
	configs := []Config{
		{ModelFamily: ModelFamilyDavidson, InitialRating: 1500, KFactor: 20, HomeAdvantage: 60, DrawFactor: 0.9, Scale: 400},
		{ModelFamily: ModelFamilyDrawDecay, InitialRating: 1500, KFactor: 20, HomeAdvantage: 70, BaseDraw: 0.4, DrawScale: 75, Scale: 400},
	}
	for _, cfg := range configs {
		probs := predict(1500, 1500, cfg)
		sum := probs.HomeWin + probs.Draw + probs.AwayWin
		if math.Abs(sum-1.0) > 1e-9 {
			t.Fatalf("family %s probabilities sum = %.12f, want 1", cfg.ModelFamily, sum)
		}
	}
}

func TestRunMatchesDeterministic(t *testing.T) {
	matches := []Match{
		{Season: "1", HomeTeam: "A", AwayTeam: "B", Result: "H"},
		{Season: "1", HomeTeam: "B", AwayTeam: "A", Result: "D"},
	}
	cfg := Config{ModelFamily: ModelFamilyDrawDecay, InitialRating: 1500, KFactor: 28, HomeAdvantage: 70, BaseDraw: 0.4, DrawScale: 75, Scale: 400}
	m1, r1, p1 := RunMatchesDetailed(matches, cfg, nil, true)
	m2, r2, p2 := RunMatchesDetailed(matches, cfg, nil, true)
	if m1 != m2 {
		t.Fatalf("metrics differ: %#v vs %#v", m1, m2)
	}
	if len(r1) != len(r2) || math.Abs(r1["A"]-r2["A"]) > 1e-9 || math.Abs(r1["B"]-r2["B"]) > 1e-9 {
		t.Fatalf("ratings differ: %#v vs %#v", r1, r2)
	}
	if len(p1) != len(p2) || p1[0].PredictedResult != p2[0].PredictedResult {
		t.Fatalf("predictions differ: %#v vs %#v", p1, p2)
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

func TestNaiveFrequency(t *testing.T) {
	matches := []Match{
		{Result: "H"},
		{Result: "H"},
		{Result: "D"},
		{Result: "A"},
		{Result: "H"},
	}
	model := NewNaiveFrequency(matches)
	if model.Prediction != "H" {
		t.Fatalf("prediction = %s, want H", model.Prediction)
	}
	if model.HomeRate != 0.6 || model.DrawRate != 0.2 || model.AwayRate != 0.2 {
		t.Fatalf("rates = %.2f/%.2f/%.2f, want 0.6/0.2/0.2", model.HomeRate, model.DrawRate, model.AwayRate)
	}
	metrics := EvaluateNaiveFrequency(matches, model)
	if metrics.Accuracy != 0.6 {
		t.Fatalf("accuracy = %.4f, want 0.6", metrics.Accuracy)
	}
}

func TestNaiveFrequencyEmptyMatches(t *testing.T) {
	model := NewNaiveFrequency(nil)
	if model.Prediction != "H" {
		t.Fatalf("empty prediction = %s, want H", model.Prediction)
	}
}

func TestNaiveFrequencyDrawDominant(t *testing.T) {
	matches := []Match{
		{Result: "D"},
		{Result: "D"},
		{Result: "H"},
	}
	model := NewNaiveFrequency(matches)
	if model.Prediction != "D" {
		t.Fatalf("prediction = %s, want D", model.Prediction)
	}
}

func TestNaiveFrequencyDeterministic(t *testing.T) {
	matches := []Match{
		{Result: "H"},
		{Result: "A"},
		{Result: "H"},
	}
	m1 := NewNaiveFrequency(matches)
	m2 := NewNaiveFrequency(matches)
	if m1.Prediction != m2.Prediction || m1.HomeRate != m2.HomeRate {
		t.Fatalf("non-deterministic: %#v vs %#v", m1, m2)
	}
}

func TestRunBacktestReturnsHoldoutPredictions(t *testing.T) {
	matches := []Match{
		{Season: "pre1", HomeTeam: "A", AwayTeam: "B", Result: "H", HomeGoals: 2, AwayGoals: 1},
		{Season: "pre1", HomeTeam: "C", AwayTeam: "D", Result: "A", HomeGoals: 0, AwayGoals: 1},
		{Season: "val", HomeTeam: "A", AwayTeam: "C", Result: "D", HomeGoals: 1, AwayGoals: 1},
		{Season: "hold", HomeTeam: "B", AwayTeam: "D", Result: "A", HomeGoals: 1, AwayGoals: 2},
	}
	grid := []Config{{ModelFamily: ModelFamilyDrawDecay, InitialRating: 1500, KFactor: 28, HomeAdvantage: 70, BaseDraw: 0.4, DrawScale: 75, Scale: 400}}
	report, err := RunBacktest(matches, "val", "hold", grid, []string{"sample.csv"})
	if err != nil {
		t.Fatalf("RunBacktest error: %v", err)
	}
	if report.ValidationSeason != "val" || report.HoldoutSeason != "hold" {
		t.Fatalf("unexpected seasons: %#v", report)
	}
	if len(report.HoldoutMatches) != 1 {
		t.Fatalf("holdout predictions = %d, want 1", len(report.HoldoutMatches))
	}
	if report.HoldoutMatches[0].ActualAwayGoals != 2 {
		t.Fatalf("actual away goals = %d, want 2", report.HoldoutMatches[0].ActualAwayGoals)
	}
}
