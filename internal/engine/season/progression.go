package season

import (
	"github.com/ismael/football-analytics/internal/domain"
	"github.com/ismael/football-analytics/internal/engine/match"
)

// MatchResult links a simulated fixture to its summary.
type MatchResult struct {
	Fixture domain.Fixture
	Summary match.Summary
}

// SeasonResult holds the complete output of a simulated season.
type SeasonResult struct {
	Season   domain.Season
	Fixtures []domain.Fixture
	Results  []MatchResult
	Table    []Standing
}

// RunSeason simulates all fixtures in matchday order using deterministic seeds
// derived from baseSeed XOR'd with the fixture index. Squads and clubs are
// looked up by club ID; fixtures with missing data are skipped.
func RunSeason(
	baseSeed int64,
	season domain.Season,
	fixtures []domain.Fixture,
	squads map[int64][]domain.Player,
	clubs map[int64]domain.Club,
	formation string,
) SeasonResult {
	results := make([]MatchResult, 0, len(fixtures))

	for idx, fixture := range fixtures {
		homeClub, homeOK := clubs[fixture.HomeClubID]
		awayClub, awayOK := clubs[fixture.AwayClubID]
		homeSquad, homeSquadOK := squads[fixture.HomeClubID]
		awaySquad, awaySquadOK := squads[fixture.AwayClubID]

		if !homeOK || !awayOK || !homeSquadOK || !awaySquadOK {
			continue
		}

		homeLineup, err := match.SelectLineup(homeSquad, formation, nil)
		if err != nil {
			continue
		}
		awayLineup, err := match.SelectLineup(awaySquad, formation, nil)
		if err != nil {
			continue
		}

		homePlan := match.BuildTeamPlan(homeClub, homeLineup)
		awayPlan := match.BuildTeamPlan(awayClub, awayLineup)

		seed := baseSeed ^ int64(idx+1)
		summary := match.RunTickLoop(seed, homePlan, awayPlan, 0)

		results = append(results, MatchResult{
			Fixture: fixture,
			Summary: summary,
		})
	}

	// Build domain.Match slice for table computation.
	domainMatches := make([]domain.Match, len(results))
	for i, r := range results {
		domainMatches[i] = domain.Match{
			FixtureID: r.Fixture.ID,
			HomeGoals: r.Summary.HomeGoals,
			AwayGoals: r.Summary.AwayGoals,
		}
	}

	clubSlice := make([]domain.Club, 0, len(clubs))
	for _, c := range clubs {
		clubSlice = append(clubSlice, c)
	}

	table := ComputeTable(clubSlice, fixtures, domainMatches)

	return SeasonResult{
		Season:   season,
		Fixtures: fixtures,
		Results:  results,
		Table:    table,
	}
}
