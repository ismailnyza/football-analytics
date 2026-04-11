package matchlab

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNew_initialState(t *testing.T) {
	m := New()
	if m.simulated {
		t.Fatal("simulated should be false on init")
	}
	if m.activePaneFocus != paneSetup {
		t.Fatalf("activePaneFocus = %d, want paneSetup", m.activePaneFocus)
	}
	if len(m.formations) == 0 {
		t.Fatal("formations must not be empty")
	}
	if m.homeClub.Name == "" || m.awayClub.Name == "" {
		t.Fatal("demo clubs must have names")
	}
	if len(m.homeSquad) < 11 {
		t.Fatalf("home squad = %d players, want ≥ 11", len(m.homeSquad))
	}
	if len(m.awaySquad) < 11 {
		t.Fatalf("away squad = %d players, want ≥ 11", len(m.awaySquad))
	}
}

func TestUpdate_tabSwitchesPanes(t *testing.T) {
	m := New()
	if m.activePaneFocus != paneSetup {
		t.Fatal("expected paneSetup initially")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.activePaneFocus != paneResult {
		t.Fatalf("activePaneFocus = %d, want paneResult after tab", m.activePaneFocus)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.activePaneFocus != paneSetup {
		t.Fatalf("activePaneFocus = %d, want paneSetup after second tab", m.activePaneFocus)
	}
}

func TestUpdate_formationCyclesWithArrows(t *testing.T) {
	m := New()
	initial := m.formationIdx
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if m.formationIdx == initial {
		// l key cycles right — unless only one formation, which is not the case
		t.Fatal("formation index did not advance on 'l'")
	}
	// cycle back
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if m.formationIdx != initial {
		t.Fatalf("formation index = %d, want %d after h", m.formationIdx, initial)
	}
}

func TestUpdate_seedChanges(t *testing.T) {
	m := New()
	original := m.seed
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	if m.seed != original+1 {
		t.Fatalf("seed = %d, want %d after +", m.seed, original+1)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("-")})
	if m.seed != original {
		t.Fatalf("seed = %d, want %d after -", m.seed, original)
	}
}

func TestUpdate_simulateProducesResult(t *testing.T) {
	m := New()
	if m.simulated {
		t.Fatal("should not be simulated before 's' key")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	if !m.simulated {
		t.Fatal("simulated should be true after 's' key")
	}
	if m.summary == nil {
		t.Fatal("summary must not be nil after simulation")
	}
}

func TestUpdate_resetClearsResult(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	if !m.simulated {
		t.Fatal("expected simulated after 's'")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if m.simulated {
		t.Fatal("simulated should be false after 'r'")
	}
	if m.summary != nil {
		t.Fatal("summary should be nil after reset")
	}
}

func TestUpdate_deterministic(t *testing.T) {
	m1 := New()
	m1, _ = m1.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})

	m2 := New()
	m2, _ = m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})

	if m1.summary.HomeGoals != m2.summary.HomeGoals ||
		m1.summary.AwayGoals != m2.summary.AwayGoals {
		t.Fatalf("non-deterministic: %d-%d vs %d-%d",
			m1.summary.HomeGoals, m1.summary.AwayGoals,
			m2.summary.HomeGoals, m2.summary.AwayGoals)
	}
}

func TestView_containsExpectedRegions(t *testing.T) {
	m := New()
	view := m.View(120, 40)
	for _, fragment := range []string{
		"Match Lab Setup",
		"Home Team",
		"Away Team",
		"Formation",
		"Simulate",
	} {
		if !strings.Contains(view, fragment) {
			t.Fatalf("View() missing %q", fragment)
		}
	}
}

func TestView_showsResultAfterSimulation(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	view := m.View(120, 40)
	if !strings.Contains(view, "Match Result") {
		t.Fatal("View() missing 'Match Result' after simulation")
	}
	if !strings.Contains(view, "Stats") {
		t.Fatal("View() missing 'Stats' after simulation")
	}
}

func TestView_scrollEventLog(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	// switch to result pane
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	initial := m.eventScroll
	// scroll down once
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	// Scroll may or may not advance depending on event count; just check no panic
	_ = m.View(120, 40)
	// scroll back up
	m.eventScroll = initial
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	// should not go below 0
	if m.eventScroll < 0 {
		t.Fatalf("eventScroll = %d, must not be negative", m.eventScroll)
	}
}
