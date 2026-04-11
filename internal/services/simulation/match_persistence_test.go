package simulation

import (
	"context"
	"errors"
	"testing"

	"github.com/ismael/football-analytics/internal/domain"
	matchengine "github.com/ismael/football-analytics/internal/engine/match"
)

func TestPersistMatchSummaryStoresMatchAndEvents(t *testing.T) {
	writer := &fakeMatchWriter{
		match: domain.Match{ID: 55},
	}

	record, err := PersistMatchSummary(context.Background(), writer, 10, 20, 30, matchengine.Summary{
		TotalTicks: 90,
		HomeGoals:  2,
		AwayGoals:  1,
		Events: []domain.MatchEvent{
			{Tick: 1, Minute: 0, Type: "build_up"},
			{Tick: 2, Minute: 0, Type: "goal"},
		},
	})
	if err != nil {
		t.Fatalf("PersistMatchSummary() error = %v", err)
	}

	if record.ID != 55 {
		t.Fatalf("record.ID = %d, want 55", record.ID)
	}
	if len(writer.savedEvents) != 2 {
		t.Fatalf("saved events = %d, want 2", len(writer.savedEvents))
	}
	for _, event := range writer.savedEvents {
		if event.MatchID != 55 {
			t.Fatalf("event.MatchID = %d, want 55", event.MatchID)
		}
	}
}

func TestPersistMatchSummaryPropagatesCreateErrors(t *testing.T) {
	writer := &fakeMatchWriter{createErr: errors.New("boom")}
	_, err := PersistMatchSummary(context.Background(), writer, 1, 1, 1, matchengine.Summary{})
	if err == nil {
		t.Fatal("expected create error")
	}
}

type fakeMatchWriter struct {
	match       domain.Match
	createErr   error
	saveErr     error
	savedEvents []domain.MatchEvent
}

func (f *fakeMatchWriter) CreateMatch(context.Context, domain.Match) (domain.Match, error) {
	if f.createErr != nil {
		return domain.Match{}, f.createErr
	}
	return f.match, nil
}

func (f *fakeMatchWriter) SaveMatchEvents(_ context.Context, events []domain.MatchEvent) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.savedEvents = append(f.savedEvents, events...)
	return nil
}
