package simulation

import (
	"context"
	"fmt"
	"time"

	"github.com/ismael/football-analytics/internal/domain"
	matchengine "github.com/ismael/football-analytics/internal/engine/match"
)

// MatchWriter persists match summaries and event logs.
type MatchWriter interface {
	CreateMatch(ctx context.Context, match domain.Match) (domain.Match, error)
	SaveMatchEvents(ctx context.Context, events []domain.MatchEvent) error
}

// PersistMatchSummary stores a simulated match and its event log through the repository layer.
func PersistMatchSummary(
	ctx context.Context,
	writer MatchWriter,
	fixtureID int64,
	branchID int64,
	seed int64,
	summary matchengine.Summary,
) (domain.Match, error) {
	record, err := writer.CreateMatch(ctx, domain.Match{
		FixtureID:   fixtureID,
		BranchID:    branchID,
		Seed:        seed,
		HomeGoals:   summary.HomeGoals,
		AwayGoals:   summary.AwayGoals,
		TickCount:   summary.TotalTicks,
		SimulatedAt: time.Now().UTC(),
		Status:      "completed",
	})
	if err != nil {
		return domain.Match{}, fmt.Errorf("create match: %w", err)
	}

	events := make([]domain.MatchEvent, 0, len(summary.Events))
	for _, event := range summary.Events {
		event.MatchID = record.ID
		events = append(events, event)
	}
	if err := writer.SaveMatchEvents(ctx, events); err != nil {
		return domain.Match{}, fmt.Errorf("save match events: %w", err)
	}

	return record, nil
}
