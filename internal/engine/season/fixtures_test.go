package season

import (
	"fmt"
	"testing"
	"time"

	"github.com/ismael/football-analytics/internal/domain"
)

func makeClubs(ids ...int64) []domain.Club {
	clubs := make([]domain.Club, len(ids))
	for i, id := range ids {
		clubs[i] = domain.Club{ID: id, Name: fmt.Sprintf("Club%d", id), ShortName: fmt.Sprintf("C%d", id)}
	}
	return clubs
}

func TestGenerateFixtures_errorOnTooFewClubs(t *testing.T) {
	season := domain.Season{ID: 1}
	_, err := GenerateFixtures(season, []domain.Club{{ID: 1}}, time.Time{})
	if err == nil {
		t.Fatal("expected error for fewer than 2 clubs")
	}
}

func TestGenerateFixtures_evenCount(t *testing.T) {
	clubs := makeClubs(1, 2, 3, 4)
	season := domain.Season{ID: 1}
	fixtures, err := GenerateFixtures(season, clubs, time.Time{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 4 clubs → 4*3 = 12 fixtures
	if len(fixtures) != 12 {
		t.Fatalf("fixture count = %d, want 12", len(fixtures))
	}
}

func TestGenerateFixtures_oddCount(t *testing.T) {
	clubs := makeClubs(1, 2, 3)
	season := domain.Season{ID: 1}
	fixtures, err := GenerateFixtures(season, clubs, time.Time{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 3 clubs → padded to 4 → 4*3 = 12 max but byes remove 3+3 = 6 → 6 fixtures
	if len(fixtures) != 6 {
		t.Fatalf("fixture count = %d, want 6 for 3 clubs", len(fixtures))
	}
}

func TestGenerateFixtures_matchdayRange(t *testing.T) {
	clubs := makeClubs(1, 2, 3, 4)
	season := domain.Season{ID: 1}
	fixtures, err := GenerateFixtures(season, clubs, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	seen := map[int]bool{}
	for _, f := range fixtures {
		seen[f.Matchday] = true
	}
	// 4 clubs → 6 matchdays (3 + 3 return legs)
	if len(seen) != 6 {
		t.Fatalf("matchday count = %d, want 6", len(seen))
	}
	for md := 1; md <= 6; md++ {
		if !seen[md] {
			t.Fatalf("matchday %d missing", md)
		}
	}
}

func TestGenerateFixtures_noSelfFixtures(t *testing.T) {
	clubs := makeClubs(1, 2, 3, 4, 5, 6)
	season := domain.Season{ID: 1}
	fixtures, err := GenerateFixtures(season, clubs, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fixtures {
		if f.HomeClubID == f.AwayClubID {
			t.Fatalf("self-fixture found for club %d", f.HomeClubID)
		}
	}
}

func TestGenerateFixtures_allPairsPresent(t *testing.T) {
	clubs := makeClubs(10, 20, 30, 40)
	season := domain.Season{ID: 1}
	fixtures, err := GenerateFixtures(season, clubs, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	// Every ordered pair (a,b) where a≠b should appear exactly once.
	type pair struct{ home, away int64 }
	counts := map[pair]int{}
	for _, f := range fixtures {
		counts[pair{f.HomeClubID, f.AwayClubID}]++
	}
	ids := []int64{10, 20, 30, 40}
	for _, h := range ids {
		for _, a := range ids {
			if h == a {
				continue
			}
			if counts[pair{h, a}] != 1 {
				t.Fatalf("pair (%d,%d) appears %d times, want 1", h, a, counts[pair{h, a}])
			}
		}
	}
}

func TestGenerateFixtures_seasonIDPropagated(t *testing.T) {
	clubs := makeClubs(1, 2)
	season := domain.Season{ID: 99}
	fixtures, err := GenerateFixtures(season, clubs, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fixtures {
		if f.SeasonID != 99 {
			t.Fatalf("fixture seasonID = %d, want 99", f.SeasonID)
		}
	}
}

func TestGenerateFixtures_scheduledDates(t *testing.T) {
	clubs := makeClubs(1, 2, 3, 4)
	season := domain.Season{ID: 1}
	start := time.Date(2024, 8, 10, 0, 0, 0, 0, time.UTC)
	fixtures, err := GenerateFixtures(season, clubs, start)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range fixtures {
		if f.Scheduled.IsZero() {
			t.Fatalf("fixture on matchday %d has zero scheduled date", f.Matchday)
		}
		expected := start.AddDate(0, 0, (f.Matchday-1)*7)
		if !f.Scheduled.Equal(expected) {
			t.Fatalf("fixture matchday %d scheduled %v, want %v", f.Matchday, f.Scheduled, expected)
		}
	}
}

func TestGenerateFixtures_returnLegsSwapped(t *testing.T) {
	clubs := makeClubs(1, 2, 3, 4)
	season := domain.Season{ID: 1}
	fixtures, err := GenerateFixtures(season, clubs, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	// Split into first and second half.
	type pair struct{ home, away int64 }
	first := map[pair]bool{}
	second := map[pair]bool{}
	rounds := len(clubs) - 1 // 3 for 4 clubs
	for _, f := range fixtures {
		if f.Matchday <= rounds {
			first[pair{f.HomeClubID, f.AwayClubID}] = true
		} else {
			second[pair{f.HomeClubID, f.AwayClubID}] = true
		}
	}
	for p := range first {
		reversed := pair{p.away, p.home}
		if !second[reversed] {
			t.Fatalf("return leg missing for fixture (%d,%d)", p.home, p.away)
		}
	}
}
