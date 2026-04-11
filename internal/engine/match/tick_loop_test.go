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
		switch event.Type {
		case "build_up", "penetration", "shot", "goal", "save", "block", "turnover", "injury", "yellow", "red", "corner", "free_kick", "penalty", "clearance":
		default:
			t.Fatalf("unexpected event type %q", event.Type)
		}
	}
}

func TestRunTickLoopProducesOrderedActionPhases(t *testing.T) {
	home, away := samplePlans()

	summary := RunTickLoop(9, home, away, 40)
	lastTick := -1
	seenBuildUp := false
	for _, event := range summary.Events {
		if event.Tick < lastTick {
			t.Fatalf("events out of order: %d before %d", event.Tick, lastTick)
		}
		lastTick = event.Tick
		if event.Type == "build_up" {
			seenBuildUp = true
		}
	}
	if !seenBuildUp {
		t.Fatal("expected at least one build_up event")
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

func TestRunTickLoopAccumulatesFatigue(t *testing.T) {
	home, away := samplePlans()

	summary := RunTickLoop(11, home, away, 90)
	if summary.HomeAverageFatigue <= 0 {
		t.Fatalf("HomeAverageFatigue = %f, want > 0", summary.HomeAverageFatigue)
	}
	if summary.AwayAverageFatigue <= 0 {
		t.Fatalf("AwayAverageFatigue = %f, want > 0", summary.AwayAverageFatigue)
	}
}

func TestFatigueAdjustedPlanReducesStrength(t *testing.T) {
	assignments, err := SelectLineup(sampleSquad(), "4-3-3", nil)
	if err != nil {
		t.Fatalf("SelectLineup() error = %v", err)
	}

	base := BuildTeamPlan(domain.Club{Name: "Arsenal", ShortName: "ARS"}, assignments)
	fatigue := initialFatigue(assignments)
	for i := 0; i < 300; i++ {
		applyFatigueTick(fatigue, assignments, true)
	}

	adjusted := fatigueAdjustedPlan(base, fatigue, 0)
	if adjusted.Attack >= base.Attack {
		t.Fatalf("adjusted attack = %d, want less than %d", adjusted.Attack, base.Attack)
	}
	if adjusted.Control >= base.Control {
		t.Fatalf("adjusted control = %d, want less than %d", adjusted.Control, base.Control)
	}
}

func TestRunTickLoopInjuriesAreDeterministic(t *testing.T) {
	home, away := samplePlans()

	first := RunTickLoop(18, home, away, 180)
	second := RunTickLoop(18, home, away, 180)
	if !reflect.DeepEqual(first.Injuries, second.Injuries) {
		t.Fatal("injury summaries differ for identical inputs")
	}
}

func TestMaybeInjureTeamTriggersForHighFatigueCandidate(t *testing.T) {
	assignments, err := SelectLineup(sampleSquad(), "4-3-3", nil)
	if err != nil {
		t.Fatalf("SelectLineup() error = %v", err)
	}
	plan := BuildTeamPlan(domain.Club{Name: "Home", ShortName: "HOM"}, assignments)
	fatigue := initialFatigue(assignments)
	for _, assignment := range assignments {
		if assignment.Player.PrimaryPosition != domain.PositionGK {
			fatigue[assignment.Player.ID] = 4.0
		}
	}
	injured := make(map[int64]Injury)

	injury, ok := maybeInjureTeam(18, 24, plan, fatigue, injured)
	if !ok {
		t.Fatal("expected injury trigger for deterministic high-fatigue input")
	}
	if injury.Team != "HOM" {
		t.Fatalf("injury.Team = %q, want HOM", injury.Team)
	}
	if injury.Severity == "" {
		t.Fatal("injury severity is empty")
	}
}

func TestRunTickLoopCardsAreDeterministic(t *testing.T) {
	home, away := samplePlans()

	first := RunTickLoop(24, home, away, 180)
	second := RunTickLoop(24, home, away, 180)
	if !reflect.DeepEqual(first.Cards, second.Cards) {
		t.Fatal("card summaries differ for identical inputs")
	}
}

func TestMaybeCardTeamTriggersForHighFatigueCandidate(t *testing.T) {
	assignments, err := SelectLineup(sampleSquad(), "4-3-3", nil)
	if err != nil {
		t.Fatalf("SelectLineup() error = %v", err)
	}
	plan := BuildTeamPlan(domain.Club{Name: "Home", ShortName: "HOM"}, assignments)
	fatigue := initialFatigue(assignments)
	for _, assignment := range assignments {
		if assignment.Player.PrimaryPosition != domain.PositionGK {
			fatigue[assignment.Player.ID] = 4.1
		}
	}

	record, ok := maybeCardTeam(11, 16, plan, fatigue, map[int64]int{}, map[int64]CardRecord{}, map[int64]Injury{})
	if !ok {
		t.Fatal("expected card trigger for deterministic high-fatigue input")
	}
	if record.Team != "HOM" {
		t.Fatalf("record.Team = %q, want HOM", record.Team)
	}
	if record.Card == "" {
		t.Fatal("record card is empty")
	}
}

func TestResolveSetPieceTypeCanTriggerDeterministically(t *testing.T) {
	home, away := samplePlans()

	for seed := int64(1); seed <= 20; seed++ {
		for tick := 1; tick <= 40; tick++ {
			setPiece, ok := resolveSetPieceType(seed, tick, home, away)
			if !ok {
				continue
			}
			switch setPiece {
			case "corner", "free_kick", "penalty":
				return
			default:
				t.Fatalf("unexpected set piece %q", setPiece)
			}
		}
	}
	t.Fatal("expected at least one deterministic set-piece trigger")
}

func TestSummaryAggregatesStatsFromEvents(t *testing.T) {
	home, away := samplePlans()

	summary := RunTickLoop(24, home, away, 180)
	if summary.Stats.Goals != summary.HomeGoals+summary.AwayGoals {
		t.Fatalf("stats goals = %d, want %d", summary.Stats.Goals, summary.HomeGoals+summary.AwayGoals)
	}
	if summary.Stats.BuildUps == 0 {
		t.Fatal("expected at least one build-up event in stats")
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
