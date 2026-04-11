package matchlab

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ismael/football-analytics/internal/app"
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

func TestNewWithConfig_usesStorageBackedTeams(t *testing.T) {
	cfg, closeStore, err := app.OpenLocalStore(context.Background(), app.Config{StateDir: t.TempDir()})
	if err != nil {
		t.Fatalf("OpenLocalStore() error = %v", err)
	}
	defer closeStore()

	m := NewWithConfig(cfg)
	if len(m.homeSquad) < 11 {
		t.Fatalf("homeSquad len = %d, want at least 11", len(m.homeSquad))
	}
	if len(m.awaySquad) < 11 {
		t.Fatalf("awaySquad len = %d, want at least 11", len(m.awaySquad))
	}
	if m.homeClub.ID == 0 || m.awayClub.ID == 0 {
		t.Fatal("expected storage-backed club IDs")
	}
}

func TestUpdate_teamCyclingChangesSelectedClubs(t *testing.T) {
	cfg, closeStore, err := app.OpenLocalStore(context.Background(), app.Config{StateDir: t.TempDir()})
	if err != nil {
		t.Fatalf("OpenLocalStore() error = %v", err)
	}
	defer closeStore()

	m := NewWithConfig(cfg)
	originalHome := m.homeClub.ID
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]")})
	if m.homeClub.ID == originalHome {
		t.Fatal("expected home club to change after ]")
	}
}

func TestUpdate_pLoadsSavedMatchReplay(t *testing.T) {
	cfg, closeStore, err := app.OpenLocalStore(context.Background(), app.Config{StateDir: t.TempDir()})
	if err != nil {
		t.Fatalf("OpenLocalStore() error = %v", err)
	}
	defer closeStore()

	m := NewWithConfig(cfg)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	if len(m.history) == 0 {
		t.Fatal("expected saved match history after simulate")
	}
	m.simulated = false
	m.summary = nil
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("p")})
	if !m.simulated || m.summary == nil {
		t.Fatal("expected replay to load after p")
	}
}

func TestUpdate_cCreatesBranchFromSavedMatch(t *testing.T) {
	cfg, closeStore, err := app.OpenLocalStore(context.Background(), app.Config{StateDir: t.TempDir()})
	if err != nil {
		t.Fatalf("OpenLocalStore() error = %v", err)
	}
	defer closeStore()

	originalBranchID := cfg.State.ActiveBranchID
	m := NewWithConfig(cfg)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	if len(m.history) == 0 {
		t.Fatal("expected match history after simulate")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	if cfg.State.ActiveBranchID == originalBranchID {
		t.Fatal("expected shared active branch to change after c")
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

// ---------------------------------------------------------------------------
// SEC-012: result detail and event log view tests
// ---------------------------------------------------------------------------

func TestUpdate_vKeyOpensResultView(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	if !m.simulated {
		t.Fatal("expected simulated after s")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	if m.viewMode != viewModeResult {
		t.Fatalf("viewMode = %d, want viewModeResult after v", m.viewMode)
	}
}

func TestUpdate_vKeyIgnoredWhenNotSimulated(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	if m.viewMode != viewModeSplit {
		t.Fatalf("viewMode = %d, want viewModeSplit when not simulated", m.viewMode)
	}
}

func TestUpdate_eKeyOpensEventLog(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if m.viewMode != viewModeEvents {
		t.Fatalf("viewMode = %d, want viewModeEvents after e", m.viewMode)
	}
}

func TestUpdate_eKeyIgnoredWhenNotSimulated(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if m.viewMode != viewModeSplit {
		t.Fatalf("viewMode = %d, want viewModeSplit when not simulated", m.viewMode)
	}
}

func TestUpdate_bKeyReturnsToSplitFromResult(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	if m.viewMode != viewModeResult {
		t.Fatal("expected viewModeResult before b")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if m.viewMode != viewModeSplit {
		t.Fatalf("viewMode = %d, want viewModeSplit after b", m.viewMode)
	}
}

func TestUpdate_bKeyReturnsToSplitFromEvents(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if m.viewMode != viewModeSplit {
		t.Fatalf("viewMode = %d, want viewModeSplit after b", m.viewMode)
	}
}

func TestUpdate_switchBetweenResultAndEvents(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	if m.viewMode != viewModeResult {
		t.Fatal("expected viewModeResult")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if m.viewMode != viewModeEvents {
		t.Fatal("expected viewModeEvents after e from result view")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	if m.viewMode != viewModeResult {
		t.Fatal("expected viewModeResult after v from events view")
	}
}

func TestUpdate_filterCyclesInEventLog(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	if m.eventFilter != filterAll {
		t.Fatal("expected filterAll initially")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	if m.eventFilter == filterAll {
		t.Fatal("eventFilter should advance after f")
	}
	// cycle through all filters back to filterAll
	for i := 0; i < len(filterLabels)-1; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("f")})
	}
	if m.eventFilter != filterAll {
		t.Fatalf("eventFilter = %d, want filterAll after full cycle", m.eventFilter)
	}
}

func TestView_resultViewContainsMatchData(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	view := m.View(120, 40)
	for _, fragment := range []string{
		"Match Result",
		"Statistics",
		"Team Condition",
		"ARS",
		"LIV",
	} {
		if !strings.Contains(view, fragment) {
			t.Fatalf("result view missing %q", fragment)
		}
	}
}

func TestView_eventLogViewContainsFilterBar(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("e")})
	view := m.View(120, 40)
	for _, fragment := range []string{
		"Event Log",
		"All",
		"Goals",
		"Cards",
		"Injuries",
	} {
		if !strings.Contains(view, fragment) {
			t.Fatalf("event log view missing %q", fragment)
		}
	}
}

func TestView_resultScrollBoundsCheck(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	// scroll down many times
	for i := 0; i < 100; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	}
	// scroll back up past zero
	for i := 0; i < 200; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	}
	if m.resultScroll < 0 {
		t.Fatalf("resultScroll = %d, must not be negative", m.resultScroll)
	}
	// View must not panic
	_ = m.View(120, 40)
}

func TestView_splitViewHintsAfterSimulation(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	view := m.View(120, 40)
	if !strings.Contains(view, "v: result detail") {
		t.Fatal("split view should hint 'v: result detail' after simulation")
	}
	if !strings.Contains(view, "e: event log") {
		t.Fatal("split view should hint 'e: event log' after simulation")
	}
}
