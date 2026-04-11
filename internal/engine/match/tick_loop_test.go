package match

import (
	"reflect"
	"testing"

	"github.com/ismael/football-analytics/internal/domain"
)

func TestRunTickLoopIsDeterministic(t *testing.T) {
	home, away := samplePlans()

	first := RunTickLoop(42, home, away, 60)
	second := RunTickLoop(42, home, away, 60)

	if !reflect.DeepEqual(first, second) {
		t.Fatal("RunTickLoop() is not deterministic for identical inputs")
	}
}

func TestRunTickLoopTracksTickCountAndEvents(t *testing.T) {
	home, away := samplePlans()

	summary := RunTickLoop(7, home, away, 30)
	if summary.TotalTicks != 30 {
		t.Fatalf("TotalTicks = %d, want 30", summary.TotalTicks)
	}
	for _, event := range summary.Events {
		if event.Tick <= 0 || event.Tick > 30 {
			t.Fatalf("event tick = %d out of range", event.Tick)
		}
		if event.Type != "chance" && event.Type != "goal" {
			t.Fatalf("unexpected event type %q", event.Type)
		}
	}
}

func TestBuildTeamPlanDerivesStableRatings(t *testing.T) {
	assignments, err := SelectLineup(sampleSquad(), "4-3-3", nil)
	if err != nil {
		t.Fatalf("SelectLineup() error = %v", err)
	}

	plan := BuildTeamPlan(domain.Club{Name: "Arsenal", ShortName: "ARS"}, assignments)
	if plan.Attack == 0 || plan.Control == 0 || plan.Defence == 0 {
		t.Fatalf("invalid plan ratings: %#v", plan)
	}
}

func samplePlans() (TeamPlan, TeamPlan) {
	homeAssignments, _ := SelectLineup(sampleSquad(), "4-3-3", nil)
	awaySquad := append([]domain.Player(nil), sampleSquad()...)
	for idx := range awaySquad {
		awaySquad[idx].ID += 100
		awaySquad[idx].FirstName = "Away" + awaySquad[idx].FirstName
	}
	awayAssignments, _ := SelectLineup(awaySquad, "4-3-3", nil)

	home := BuildTeamPlan(domain.Club{Name: "Home", ShortName: "HOM"}, homeAssignments)
	away := BuildTeamPlan(domain.Club{Name: "Away", ShortName: "AWY"}, awayAssignments)
	return home, away
}
