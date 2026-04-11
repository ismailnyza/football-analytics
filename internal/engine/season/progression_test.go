package season

import (
	"testing"
	"time"

	"github.com/ismael/football-analytics/internal/domain"
	"github.com/ismael/football-analytics/internal/engine/match"
)

func noDate() time.Time { return time.Time{} }

func buildDemoSquad(base int64) []domain.Player {
	type spec struct {
		pos domain.Position
		ovr int
	}
	specs := []spec{
		{domain.PositionGK, 80}, {domain.PositionGK, 70},
		{domain.PositionRB, 75}, {domain.PositionCB, 80}, {domain.PositionCB, 78}, {domain.PositionLB, 75},
		{domain.PositionDM, 78}, {domain.PositionCM, 80}, {domain.PositionCM, 77},
		{domain.PositionRW, 79}, {domain.PositionLW, 78},
		{domain.PositionST, 82}, {domain.PositionST, 75},
		{domain.PositionAM, 76}, {domain.PositionCB, 73},
		{domain.PositionRB, 70}, {domain.PositionLB, 70},
		{domain.PositionDM, 72}, {domain.PositionCM, 71},
		{domain.PositionST, 70},
	}
	players := make([]domain.Player, len(specs))
	for i, s := range specs {
		cid := base
		players[i] = domain.Player{
			ID:              base*100 + int64(i+1),
			ClubID:          &cid,
			PrimaryPosition: s.pos,
			Attributes: domain.PlayerAttributes{
				Overall:   s.ovr,
				Pace:      s.ovr - 3,
				Shooting:  s.ovr - 5,
				Passing:   s.ovr - 4,
				Defending: s.ovr - 6,
				Keeping:   func() int { if s.pos == domain.PositionGK { return s.ovr + 5 }; return 40 }(),
			},
		}
	}
	return players
}

func TestRunSeason_twoClubs(t *testing.T) {
	clubs := map[int64]domain.Club{
		1: {ID: 1, Name: "Home FC", ShortName: "HFC"},
		2: {ID: 2, Name: "Away FC", ShortName: "AFC"},
	}
	squads := map[int64][]domain.Player{
		1: buildDemoSquad(1),
		2: buildDemoSquad(2),
	}
	season := domain.Season{ID: 1}
	clubSlice := makeClubs(1, 2)
	clubs[1] = domain.Club{ID: 1, Name: "Home FC", ShortName: "HFC"}
	clubs[2] = domain.Club{ID: 2, Name: "Away FC", ShortName: "AFC"}

	fixtures, err := GenerateFixtures(season, clubSlice, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	// Assign fixture IDs for table lookup.
	for i := range fixtures {
		fixtures[i].ID = int64(i + 1)
	}

	result := RunSeason(42, season, fixtures, squads, clubs, "4-3-3")
	if len(result.Results) != 2 {
		t.Fatalf("expected 2 match results, got %d", len(result.Results))
	}
	if len(result.Table) != 2 {
		t.Fatalf("expected 2 standings, got %d", len(result.Table))
	}
	// Total goals across all matches must be non-negative.
	totalGoals := 0
	for _, r := range result.Results {
		totalGoals += r.Summary.HomeGoals + r.Summary.AwayGoals
	}
	if totalGoals < 0 {
		t.Fatal("negative total goals")
	}
}

func TestRunSeason_deterministic(t *testing.T) {
	clubs := map[int64]domain.Club{
		1: {ID: 1, Name: "A", ShortName: "A"},
		2: {ID: 2, Name: "B", ShortName: "B"},
	}
	squads := map[int64][]domain.Player{
		1: buildDemoSquad(1),
		2: buildDemoSquad(2),
	}
	season := domain.Season{ID: 1}
	fixtures, _ := GenerateFixtures(season, makeClubs(1, 2), time.Time{})
	for i := range fixtures {
		fixtures[i].ID = int64(i + 1)
	}

	r1 := RunSeason(42, season, fixtures, squads, clubs, "4-3-3")
	r2 := RunSeason(42, season, fixtures, squads, clubs, "4-3-3")

	for i := range r1.Results {
		if r1.Results[i].Summary.HomeGoals != r2.Results[i].Summary.HomeGoals ||
			r1.Results[i].Summary.AwayGoals != r2.Results[i].Summary.AwayGoals {
			t.Fatalf("non-deterministic results at index %d", i)
		}
	}
}

func TestRunSeason_skipsMissingSquad(t *testing.T) {
	clubs := map[int64]domain.Club{
		1: {ID: 1, Name: "A", ShortName: "A"},
		2: {ID: 2, Name: "B", ShortName: "B"},
	}
	// Only one squad provided.
	squads := map[int64][]domain.Player{
		1: buildDemoSquad(1),
	}
	season := domain.Season{ID: 1}
	fixtures, _ := GenerateFixtures(season, makeClubs(1, 2), time.Time{})
	for i := range fixtures {
		fixtures[i].ID = int64(i + 1)
	}
	result := RunSeason(42, season, fixtures, squads, clubs, "4-3-3")
	if len(result.Results) != 0 {
		t.Fatalf("expected 0 results when squad missing, got %d", len(result.Results))
	}
}

func TestRecovery_fatigueReduces(t *testing.T) {
	fatigue := map[int64]float64{1: 3.0, 2: 1.0, 3: 0.5}
	params := DefaultRecoveryParams()
	recovered := ApplyFatigueRecovery(fatigue, params)
	if recovered[1] >= fatigue[1] {
		t.Fatal("fatigue should decrease after recovery")
	}
	if recovered[3] < 0 {
		t.Fatal("fatigue must not go below 0")
	}
}

func TestRecovery_filterActiveInjuries(t *testing.T) {
	injuries := []match.Injury{
		{PlayerID: 1, Severity: "minor"},   // heals in 7 days
		{PlayerID: 2, Severity: "moderate"}, // heals in 21 days
		{PlayerID: 3, Severity: "major"},    // heals in 42 days
	}

	// After 7 days: minor healed
	active := FilterActiveInjuries(injuries, 7)
	if len(active) != 2 {
		t.Fatalf("after 7 days: expected 2 active, got %d", len(active))
	}

	// After 21 days: minor + moderate healed
	active = FilterActiveInjuries(injuries, 21)
	if len(active) != 1 {
		t.Fatalf("after 21 days: expected 1 active, got %d", len(active))
	}
	if active[0].Severity != "major" {
		t.Fatal("expected only major injury to remain")
	}

	// After 42 days: all healed
	active = FilterActiveInjuries(injuries, 42)
	if len(active) != 0 {
		t.Fatalf("after 42 days: expected 0 active, got %d", len(active))
	}
}

