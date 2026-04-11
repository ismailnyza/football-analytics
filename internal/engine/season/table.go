package season

import (
	"slices"

	"github.com/ismael/football-analytics/internal/domain"
)

// Standing holds a club's accumulated league table statistics.
type Standing struct {
	ClubID         int64
	ClubName       string
	Played         int
	Won            int
	Drawn          int
	Lost           int
	GoalsFor       int
	GoalsAgainst   int
	GoalDifference int
	Points         int
}

// ComputeTable builds league standings from completed matches.
// Only matches whose FixtureID maps to a fixture in the provided slice are
// included. The returned slice is sorted by points (desc), goal difference
// (desc), goals for (desc), then club name (asc).
func ComputeTable(clubs []domain.Club, fixtures []domain.Fixture, matches []domain.Match) []Standing {
	nameByID := make(map[int64]string, len(clubs))
	for _, c := range clubs {
		nameByID[c.ID] = c.Name
	}

	// Map fixture ID → fixture for home/away club lookup.
	fixtureByID := make(map[int64]domain.Fixture, len(fixtures))
	for _, f := range fixtures {
		fixtureByID[f.ID] = f
	}

	standings := make(map[int64]*Standing, len(clubs))
	for _, c := range clubs {
		standings[c.ID] = &Standing{ClubID: c.ID, ClubName: c.Name}
	}

	for _, m := range matches {
		f, ok := fixtureByID[m.FixtureID]
		if !ok {
			continue
		}
		home, hOK := standings[f.HomeClubID]
		away, aOK := standings[f.AwayClubID]
		if !hOK || !aOK {
			continue
		}

		home.Played++
		away.Played++
		home.GoalsFor += m.HomeGoals
		home.GoalsAgainst += m.AwayGoals
		away.GoalsFor += m.AwayGoals
		away.GoalsAgainst += m.HomeGoals

		switch {
		case m.HomeGoals > m.AwayGoals:
			home.Won++
			home.Points += 3
			away.Lost++
		case m.HomeGoals < m.AwayGoals:
			away.Won++
			away.Points += 3
			home.Lost++
		default:
			home.Drawn++
			away.Drawn++
			home.Points++
			away.Points++
		}
	}

	table := make([]Standing, 0, len(standings))
	for _, s := range standings {
		s.GoalDifference = s.GoalsFor - s.GoalsAgainst
		table = append(table, *s)
	}

	slices.SortFunc(table, func(a, b Standing) int {
		if a.Points != b.Points {
			return b.Points - a.Points
		}
		if a.GoalDifference != b.GoalDifference {
			return b.GoalDifference - a.GoalDifference
		}
		if a.GoalsFor != b.GoalsFor {
			return b.GoalsFor - a.GoalsFor
		}
		if a.ClubName < b.ClubName {
			return -1
		}
		if a.ClubName > b.ClubName {
			return 1
		}
		return 0
	})

	return table
}
