// Package season provides fixture generation, table computation, and season
// progression logic for the football simulation engine.
package season

import (
	"fmt"
	"time"

	"github.com/ismael/football-analytics/internal/domain"
)

// GenerateFixtures creates a full home-and-away round-robin schedule for the
// provided clubs within the given season. Matchday numbers start at 1.
//
// If startDate is non-zero, fixtures are spaced one week apart beginning on
// that date. With N clubs there are 2*(N-1) matchdays each containing N/2
// matches. An odd club count receives a bye slot that is silently skipped.
func GenerateFixtures(season domain.Season, clubs []domain.Club, startDate time.Time) ([]domain.Fixture, error) {
	n := len(clubs)
	if n < 2 {
		return nil, fmt.Errorf("fixture generation requires at least 2 clubs, got %d", n)
	}

	// Build ID slot list; pad to even count with a bye (ID 0).
	slots := make([]int64, n)
	for i, c := range clubs {
		slots[i] = c.ID
	}
	if n%2 != 0 {
		slots = append(slots, 0)
		n++
	}

	rounds := n - 1
	matchesPerRound := n / 2

	capacity := rounds * matchesPerRound * 2
	fixtures := make([]domain.Fixture, 0, capacity)

	current := make([]int64, n)
	copy(current, slots)

	// First half: N-1 rounds.
	for round := 0; round < rounds; round++ {
		matchday := round + 1
		var scheduled time.Time
		if !startDate.IsZero() {
			scheduled = startDate.AddDate(0, 0, round*7)
		}
		for i := 0; i < matchesPerRound; i++ {
			home := current[i]
			away := current[n-1-i]
			if home == 0 || away == 0 {
				continue // bye
			}
			fixtures = append(fixtures, domain.Fixture{
				SeasonID:   season.ID,
				Matchday:   matchday,
				HomeClubID: home,
				AwayClubID: away,
				Scheduled:  scheduled,
			})
		}
		// Rotate: current[0] is pinned; rotate current[1..n-1] right by one.
		last := current[n-1]
		copy(current[2:], current[1:n-1])
		current[1] = last
	}

	// Second half: reverse home/away for return legs.
	firstHalfLen := len(fixtures)
	for i := 0; i < firstHalfLen; i++ {
		f := fixtures[i]
		returnMatchday := f.Matchday + rounds
		var scheduled time.Time
		if !startDate.IsZero() {
			scheduled = startDate.AddDate(0, 0, (returnMatchday-1)*7)
		}
		fixtures = append(fixtures, domain.Fixture{
			SeasonID:   f.SeasonID,
			Matchday:   returnMatchday,
			HomeClubID: f.AwayClubID,
			AwayClubID: f.HomeClubID,
			Scheduled:  scheduled,
		})
	}

	return fixtures, nil
}
