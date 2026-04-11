package season

import (
	"testing"

	"github.com/ismael/football-analytics/internal/domain"
)

func TestComputeTable_emptyMatches(t *testing.T) {
	clubs := makeClubs(1, 2, 3)
	table := ComputeTable(clubs, nil, nil)
	if len(table) != 3 {
		t.Fatalf("table len = %d, want 3", len(table))
	}
	for _, s := range table {
		if s.Points != 0 || s.Played != 0 {
			t.Fatalf("expected zero standing, got %+v", s)
		}
	}
}

func TestComputeTable_win(t *testing.T) {
	clubs := makeClubs(1, 2)
	fixtures := []domain.Fixture{{ID: 1, SeasonID: 1, Matchday: 1, HomeClubID: 1, AwayClubID: 2}}
	matches := []domain.Match{{FixtureID: 1, HomeGoals: 2, AwayGoals: 0}}
	table := ComputeTable(clubs, fixtures, matches)

	if table[0].ClubID != 1 || table[0].Points != 3 || table[0].Won != 1 {
		t.Fatalf("winner not on top: %+v", table[0])
	}
	if table[1].ClubID != 2 || table[1].Points != 0 || table[1].Lost != 1 {
		t.Fatalf("loser wrong: %+v", table[1])
	}
}

func TestComputeTable_draw(t *testing.T) {
	clubs := makeClubs(1, 2)
	fixtures := []domain.Fixture{{ID: 1, SeasonID: 1, Matchday: 1, HomeClubID: 1, AwayClubID: 2}}
	matches := []domain.Match{{FixtureID: 1, HomeGoals: 1, AwayGoals: 1}}
	table := ComputeTable(clubs, fixtures, matches)

	for _, s := range table {
		if s.Points != 1 || s.Drawn != 1 {
			t.Fatalf("expected 1 point each for a draw, got %+v", s)
		}
	}
}

func TestComputeTable_goalDifferenceTieBreak(t *testing.T) {
	clubs := makeClubs(1, 2, 3)
	// Club 1 wins 1-0, Club 2 wins 3-0 — both 3 pts but club 2 has better GD
	fixtures := []domain.Fixture{
		{ID: 1, SeasonID: 1, Matchday: 1, HomeClubID: 1, AwayClubID: 3},
		{ID: 2, SeasonID: 1, Matchday: 2, HomeClubID: 2, AwayClubID: 3},
	}
	matches := []domain.Match{
		{FixtureID: 1, HomeGoals: 1, AwayGoals: 0},
		{FixtureID: 2, HomeGoals: 3, AwayGoals: 0},
	}
	table := ComputeTable(clubs, fixtures, matches)
	if table[0].ClubID != 2 {
		t.Fatalf("expected club 2 first (better GD), got club %d", table[0].ClubID)
	}
	if table[1].ClubID != 1 {
		t.Fatalf("expected club 1 second, got club %d", table[1].ClubID)
	}
}

func TestComputeTable_goalsSortedCorrectly(t *testing.T) {
	clubs := makeClubs(1, 2)
	fixtures := []domain.Fixture{
		{ID: 1, SeasonID: 1, Matchday: 1, HomeClubID: 1, AwayClubID: 2},
		{ID: 2, SeasonID: 1, Matchday: 2, HomeClubID: 1, AwayClubID: 2},
	}
	matches := []domain.Match{
		{FixtureID: 1, HomeGoals: 2, AwayGoals: 1},
		{FixtureID: 2, HomeGoals: 1, AwayGoals: 0},
	}
	table := ComputeTable(clubs, fixtures, matches)
	if table[0].ClubID != 1 {
		t.Fatal("expected club 1 on top")
	}
	if table[0].GoalsFor != 3 || table[0].GoalsAgainst != 1 {
		t.Fatalf("club 1 goals wrong: %+v", table[0])
	}
	if table[1].GoalsFor != 1 || table[1].GoalsAgainst != 3 {
		t.Fatalf("club 2 goals wrong: %+v", table[1])
	}
}

func TestComputeTable_unknownFixtureIgnored(t *testing.T) {
	clubs := makeClubs(1, 2)
	// No fixtures registered — match references unknown fixture ID.
	matches := []domain.Match{{FixtureID: 99, HomeGoals: 3, AwayGoals: 0}}
	table := ComputeTable(clubs, nil, matches)
	for _, s := range table {
		if s.Played != 0 {
			t.Fatalf("expected zero played, got %+v", s)
		}
	}
}
