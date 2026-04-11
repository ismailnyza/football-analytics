package development

import (
	"testing"
	"time"

	"github.com/ismael/football-analytics/internal/domain"
)

func makePlayer(ovr, potential, age int) domain.Player {
	dob := time.Now().AddDate(-age, 0, 0)
	return domain.Player{
		DateOfBirth: dob,
		Attributes: domain.PlayerAttributes{
			Overall:   ovr,
			Potential: potential,
			Pace:      ovr,
			Shooting:  ovr,
			Passing:   ovr,
			Defending: ovr,
		},
	}
}

func TestAgePlayer_growthPhase(t *testing.T) {
	p := makePlayer(70, 85, 19)
	params := DefaultDevelopmentParams()
	updated := AgePlayer(p, 19, params)
	if updated.Attributes.Overall <= p.Attributes.Overall {
		t.Fatalf("expected growth at age 19: before=%d after=%d",
			p.Attributes.Overall, updated.Attributes.Overall)
	}
}

func TestAgePlayer_peakPhase(t *testing.T) {
	p := makePlayer(85, 88, 27)
	params := DefaultDevelopmentParams()
	updated := AgePlayer(p, 27, params)
	if updated.Attributes.Overall != p.Attributes.Overall {
		t.Fatalf("expected no change at peak age 27: before=%d after=%d",
			p.Attributes.Overall, updated.Attributes.Overall)
	}
}

func TestAgePlayer_declinePhase(t *testing.T) {
	p := makePlayer(85, 88, 33)
	params := DefaultDevelopmentParams()
	updated := AgePlayer(p, 33, params)
	if updated.Attributes.Overall >= p.Attributes.Overall {
		t.Fatalf("expected decline at age 33: before=%d after=%d",
			p.Attributes.Overall, updated.Attributes.Overall)
	}
}

func TestAgePlayer_neverBelowFloor(t *testing.T) {
	p := makePlayer(41, 50, 40)
	params := DefaultDevelopmentParams()
	for i := 0; i < 10; i++ {
		p = AgePlayer(p, 40+i, params)
	}
	if p.Attributes.Overall < 40 {
		t.Fatalf("overall = %d, must not drop below 40", p.Attributes.Overall)
	}
}

func TestShouldRetire(t *testing.T) {
	if ShouldRetire(34) {
		t.Fatal("age 34 should not trigger retirement")
	}
	if !ShouldRetire(35) {
		t.Fatal("age 35 should trigger retirement")
	}
}

func TestAgeAtDate(t *testing.T) {
	p := domain.Player{DateOfBirth: time.Date(1990, 6, 15, 0, 0, 0, 0, time.UTC)}
	on := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	age := AgeAtDate(p, on)
	if age != 34 {
		t.Fatalf("age = %d, want 34", age)
	}
}

func TestAgeAtDate_zeroDob(t *testing.T) {
	p := domain.Player{}
	age := AgeAtDate(p, time.Now())
	if age != 0 {
		t.Fatalf("age = %d, want 0 for zero DOB", age)
	}
}

func TestProcessSquadAging_retirementFiltered(t *testing.T) {
	squad := []domain.Player{
		makePlayer(80, 82, 36),
		makePlayer(75, 80, 22),
	}
	ref := time.Now()
	result := ProcessSquadAging(squad, ref, DefaultDevelopmentParams())
	if len(result.Retiring) != 1 {
		t.Fatalf("expected 1 retiring player, got %d", len(result.Retiring))
	}
	if len(result.Updated) != 1 {
		t.Fatalf("expected 1 updated player, got %d", len(result.Updated))
	}
}

func TestGenerateYouthIntake_count(t *testing.T) {
	params := DefaultYouthIntakeParams()
	players := GenerateYouthIntake(1, 1, 42, 2024, params)
	if len(players) != params.Count {
		t.Fatalf("intake count = %d, want %d", len(players), params.Count)
	}
}

func TestGenerateYouthIntake_deterministic(t *testing.T) {
	params := DefaultYouthIntakeParams()
	a := GenerateYouthIntake(1, 1, 42, 2024, params)
	b := GenerateYouthIntake(1, 1, 42, 2024, params)
	for i := range a {
		if a[i].Attributes.Overall != b[i].Attributes.Overall {
			t.Fatalf("non-deterministic at index %d", i)
		}
	}
}

func TestGenerateYouthIntake_potentialInRange(t *testing.T) {
	params := DefaultYouthIntakeParams()
	players := GenerateYouthIntake(1, 1, 42, 2024, params)
	for _, p := range players {
		if p.Attributes.Potential < params.MinPotential || p.Attributes.Potential > params.MaxPotential {
			t.Fatalf("potential %d out of range [%d, %d]",
				p.Attributes.Potential, params.MinPotential, params.MaxPotential)
		}
	}
}
