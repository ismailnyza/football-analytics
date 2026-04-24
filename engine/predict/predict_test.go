package predict

import (
	"testing"

	"github.com/ismailnyza/football-analytics/engine/baseline"
)

func TestNewPlayerPool(t *testing.T) {
	teams := map[string]float64{
		"Arsenal":   1780,
		"Liverpool": 1740,
	}
	pool := NewPlayerPool(teams)
	if len(pool.Players) != 40 {
		t.Fatalf("players = %d, want 40 (20 per team)", len(pool.Players))
	}
	if len(pool.ByTeam) != 2 {
		t.Fatalf("teams = %d, want 2", len(pool.ByTeam))
	}
	arsenalPlayers := pool.GetTeamPlayers("Arsenal")
	if len(arsenalPlayers) != 20 {
		t.Fatalf("Arsenal players = %d, want 20", len(arsenalPlayers))
	}
	if arsenalPlayers[0].Rating < arsenalPlayers[19].Rating {
		t.Fatalf("players not sorted by rating desc")
	}
}

func TestScorerProbabilities(t *testing.T) {
	teams := map[string]float64{"Arsenal": 1780}
	pool := NewPlayerPool(teams)
	preds := pool.ScorerProbabilities("Arsenal", 2)
	if len(preds) == 0 {
		t.Fatalf("no scorer predictions")
	}
	totalProb := 0.0
	for _, p := range preds {
		totalProb += p.GoalProbability
	}
	if totalProb < 1.5 || totalProb > 2.5 {
		t.Fatalf("total goal probability = %.2f, expected ~2.0", totalProb)
	}
}

func TestAssistProbabilities(t *testing.T) {
	teams := map[string]float64{"Liverpool": 1740}
	pool := NewPlayerPool(teams)
	preds := pool.AssistProbabilities("Liverpool", 2)
	if len(preds) == 0 {
		t.Fatalf("no assist predictions")
	}
}

func TestPipelinePredict(t *testing.T) {
	teams := map[string]float64{"Arsenal": 1780, "Liverpool": 1740}
	form := make(baseline.FormMap)
	cfg := baseline.Config{
		ModelFamily:   baseline.ModelFamilyDrawDecay,
		InitialRating: 1500,
		KFactor:       28,
		HomeAdvantage: 70,
		BaseDraw:      0.40,
		DrawScale:     75,
		Scale:         400,
	}
	p := NewPipeline(teams, form, cfg)
	pred, err := p.Predict("Arsenal", "Liverpool")
	if err != nil {
		t.Fatalf("predict: %v", err)
	}
	if pred.HomeTeam != "Arsenal" || pred.AwayTeam != "Liverpool" {
		t.Fatalf("teams wrong: %s vs %s", pred.HomeTeam, pred.AwayTeam)
	}
	if len(pred.HomeScorers) == 0 || len(pred.AwayScorers) == 0 {
		t.Fatalf("no scorers predicted")
	}
	if len(pred.HomeAssists) == 0 || len(pred.AwayAssists) == 0 {
		t.Fatalf("no assists predicted")
	}
	if pred.PlayerPoolSize == 0 {
		t.Fatalf("player pool empty")
	}
	if pred.PredictedScoreline == "" {
		t.Fatalf("no predicted scoreline")
	}
}

func TestPipelineDeterministic(t *testing.T) {
	teams := map[string]float64{"Arsenal": 1780, "Liverpool": 1740}
	form := make(baseline.FormMap)
	cfg := baseline.Config{
		ModelFamily:   baseline.ModelFamilyDrawDecay,
		InitialRating: 1500,
		KFactor:       28,
		HomeAdvantage: 70,
		BaseDraw:      0.40,
		DrawScale:     75,
		Scale:         400,
	}
	p := NewPipeline(teams, form, cfg)
	pred1, _ := p.Predict("Arsenal", "Liverpool")
	pred2, _ := p.Predict("Arsenal", "Liverpool")
	if pred1.PredictedResult != pred2.PredictedResult {
		t.Fatalf("non-deterministic result: %s vs %s", pred1.PredictedResult, pred2.PredictedResult)
	}
	if pred1.PredictedScoreline != pred2.PredictedScoreline {
		t.Fatalf("non-deterministic scoreline: %s vs %s", pred1.PredictedScoreline, pred2.PredictedScoreline)
	}
}

func TestReconstructMatch(t *testing.T) {
	teams := map[string]float64{"Arsenal": 1780, "Chelsea": 1620}
	pool := NewPlayerPool(teams)
	homePlayers := pool.GetTeamPlayers("Arsenal")
	awayPlayers := pool.GetTeamPlayers("Chelsea")
	homeStats := RichTeamStats{Team: "Arsenal", Shots: 14, ShotsOnTarget: 5, Fouls: 11, Corners: 6, YellowCards: 2}
	awayStats := RichTeamStats{Team: "Chelsea", Shots: 10, ShotsOnTarget: 3, Fouls: 10, Corners: 4, YellowCards: 1}
	result := ReconstructMatch("Arsenal", "Chelsea", 2, 1, homePlayers, awayPlayers, homeStats, awayStats)
	if result.Scoreline != "2-1" {
		t.Fatalf("scoreline = %s, want 2-1", result.Scoreline)
	}
	if len(result.Events) < 5 {
		t.Fatalf("events = %d, want at least 5 (3 goals + 2 cards)", len(result.Events))
	}
	goalCount := 0
	for _, e := range result.Events {
		if e.Kind == "goal" {
			goalCount++
			if e.Goal == nil || e.Goal.ScorerName == "" {
				t.Fatalf("goal event missing scorer")
			}
		}
	}
	if goalCount != 3 {
		t.Fatalf("goal events = %d, want 3", goalCount)
	}
}

func TestReconstructMatchDeterministic(t *testing.T) {
	teams := map[string]float64{"Arsenal": 1780, "Chelsea": 1620}
	pool := NewPlayerPool(teams)
	homePlayers := pool.GetTeamPlayers("Arsenal")
	awayPlayers := pool.GetTeamPlayers("Chelsea")
	homeStats := RichTeamStats{Team: "Arsenal", Shots: 14, ShotsOnTarget: 5, Fouls: 11, Corners: 6, YellowCards: 2}
	awayStats := RichTeamStats{Team: "Chelsea", Shots: 10, ShotsOnTarget: 3, Fouls: 10, Corners: 4, YellowCards: 1}
	r1 := ReconstructMatch("Arsenal", "Chelsea", 2, 1, homePlayers, awayPlayers, homeStats, awayStats)
	r2 := ReconstructMatch("Arsenal", "Chelsea", 2, 1, homePlayers, awayPlayers, homeStats, awayStats)
	if r1.Scoreline != r2.Scoreline {
		t.Fatalf("non-deterministic scoreline")
	}
	if len(r1.Events) != len(r2.Events) {
		t.Fatalf("non-deterministic event count: %d vs %d", len(r1.Events), len(r2.Events))
	}
	for i := range r1.Events {
		if r1.Events[i].Minute != r2.Events[i].Minute || r1.Events[i].Kind != r2.Events[i].Kind {
			t.Fatalf("events differ at index %d", i)
		}
	}
}
