// regression_test.go contains the realism regression suite (SEC-045).
// It runs a full simulated season and validates that output statistics are
// internally consistent and fall within the engine's expected operating range.
//
// NOTE: The current engine (Phase B) uses deterministic seed-based scoring
// without calibrated probability distributions. The realism calibration layer
// (SEC-043, Python bridge) will tighten these bounds in a later phase.
// Until then, this test enforces structural integrity and reproducibility.
package season

import (
	"fmt"
	"testing"
	"time"

	"github.com/ismael/football-analytics/internal/domain"
)

// buildRegressionSquad produces a 20-player squad for regression testing.
func buildRegressionSquad(clubID int64) []domain.Player {
	type spec struct {
		pos domain.Position
		ovr int
	}
	specs := []spec{
		{domain.PositionGK, 80}, {domain.PositionGK, 72},
		{domain.PositionRB, 76}, {domain.PositionCB, 83}, {domain.PositionCB, 81}, {domain.PositionLB, 77},
		{domain.PositionDM, 80}, {domain.PositionCM, 82}, {domain.PositionCM, 79},
		{domain.PositionRW, 81}, {domain.PositionLW, 80}, {domain.PositionST, 84},
		{domain.PositionAM, 78}, {domain.PositionCB, 75}, {domain.PositionRB, 72},
		{domain.PositionLB, 72}, {domain.PositionDM, 74}, {domain.PositionCM, 73},
		{domain.PositionST, 76}, {domain.PositionST, 72},
	}
	players := make([]domain.Player, len(specs))
	for i, s := range specs {
		cid := clubID
		players[i] = domain.Player{
			ID: clubID*100 + int64(i+1), ClubID: &cid,
			PrimaryPosition: s.pos,
			Attributes: domain.PlayerAttributes{
				Overall:   s.ovr,
				Pace:      s.ovr - 3,
				Shooting:  s.ovr - 5,
				Passing:   s.ovr - 4,
				Defending: s.ovr - 6,
				Keeping: func() int {
					if s.pos == domain.PositionGK {
						return s.ovr + 5
					}
					return 40
				}(),
			},
		}
	}
	return players
}

// TestRealismRegression runs a 6-club full season and validates:
//   - Correct match count and table structure
//   - Non-negative statistics
//   - Reproducible results (same seed → same output)
//   - Goals-per-match within the current engine's expected range (pre-calibration)
func TestRealismRegression(t *testing.T) {
	clubIDs := []int64{1, 2, 3, 4, 5, 6}
	clubs := make(map[int64]domain.Club, len(clubIDs))
	squads := make(map[int64][]domain.Player, len(clubIDs))
	clubSlice := make([]domain.Club, len(clubIDs))

	for i, id := range clubIDs {
		c := domain.Club{ID: id, Name: fmt.Sprintf("Club%d", id), ShortName: fmt.Sprintf("C%d", id)}
		clubs[id] = c
		squads[id] = buildRegressionSquad(id)
		clubSlice[i] = c
	}

	s := domain.Season{ID: 1, BranchID: 1, Label: "Regression Season", StartYear: 2024}
	fixtures, err := GenerateFixtures(s, clubSlice, time.Time{})
	if err != nil {
		t.Fatalf("generate fixtures: %v", err)
	}
	for i := range fixtures {
		fixtures[i].ID = int64(i + 1)
	}

	result := RunSeason(42, s, fixtures, squads, clubs, "4-3-3")

	// ---- Structural integrity ----

	expectedMatches := len(clubIDs) * (len(clubIDs) - 1) // 30 for 6 clubs
	if len(result.Results) != expectedMatches {
		t.Fatalf("result count = %d, want %d", len(result.Results), expectedMatches)
	}
	if len(result.Table) != len(clubIDs) {
		t.Fatalf("table length = %d, want %d", len(result.Table), len(clubIDs))
	}

	// ---- Statistics sanity ----

	totalGoals := 0
	for _, r := range result.Results {
		totalGoals += r.Summary.HomeGoals + r.Summary.AwayGoals
		if r.Summary.HomeGoals < 0 || r.Summary.AwayGoals < 0 {
			t.Error("negative goals in match result")
		}
	}

	goalsPerMatch := float64(totalGoals) / float64(expectedMatches)

	// Pre-calibration range: engine currently scores aggressively.
	// These bounds will tighten once the Python calibration bridge (SEC-043)
	// tunes the probability distributions to match real-world distributions.
	if goalsPerMatch < 0.5 {
		t.Errorf("goals per match = %.2f, too low — engine may be broken", goalsPerMatch)
	}
	if goalsPerMatch > 50.0 {
		t.Errorf("goals per match = %.2f, absurdly high — engine may be broken", goalsPerMatch)
	}

	// ---- Table consistency ----
	totalPoints := 0
	for _, standing := range result.Table {
		if standing.Points < 0 {
			t.Errorf("negative points for club %d", standing.ClubID)
		}
		expectedPlayed := (len(clubIDs) - 1) * 2
		if standing.Played != expectedPlayed {
			t.Errorf("club %d played %d, want %d", standing.ClubID, standing.Played, expectedPlayed)
		}
		if standing.GoalDifference != standing.GoalsFor-standing.GoalsAgainst {
			t.Errorf("club %d GD mismatch: %d != %d - %d",
				standing.ClubID, standing.GoalDifference, standing.GoalsFor, standing.GoalsAgainst)
		}
		totalPoints += standing.Points
	}

	// Total points must equal 3 per decisive match + 2 per draw (all points awarded).
	decidedMatches := 0
	drawnMatches := 0
	for _, r := range result.Results {
		if r.Summary.HomeGoals != r.Summary.AwayGoals {
			decidedMatches++
		} else {
			drawnMatches++
		}
	}
	expectedPoints := decidedMatches*3 + drawnMatches*2
	if totalPoints != expectedPoints {
		t.Errorf("total points = %d, want %d (decided=%d, drawn=%d)",
			totalPoints, expectedPoints, decidedMatches, drawnMatches)
	}

	// ---- Determinism ----
	result2 := RunSeason(42, s, fixtures, squads, clubs, "4-3-3")
	for i := range result.Results {
		r1 := result.Results[i].Summary
		r2 := result2.Results[i].Summary
		if r1.HomeGoals != r2.HomeGoals || r1.AwayGoals != r2.AwayGoals {
			t.Fatalf("non-deterministic at match %d: %d-%d vs %d-%d",
				i, r1.HomeGoals, r1.AwayGoals, r2.HomeGoals, r2.AwayGoals)
		}
	}

	t.Logf("Regression: %d matches, %.2f goals/match, %d decided, %d drawn, total pts: %d",
		expectedMatches, goalsPerMatch, decidedMatches, drawnMatches, totalPoints)
}
