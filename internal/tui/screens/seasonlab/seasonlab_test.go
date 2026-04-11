package seasonlab

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
	if m.mode != viewSetup {
		t.Fatalf("mode = %d, want viewSetup", m.mode)
	}
}

func TestUpdate_simulateProducesResult(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	if !m.simulated {
		t.Fatal("simulated should be true after s")
	}
	if m.result == nil {
		t.Fatal("result must not be nil")
	}
	if len(m.result.Table) == 0 {
		t.Fatal("table must not be empty")
	}
}

func TestUpdate_tKeyOpensStandings(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	if m.mode != viewStandings {
		t.Fatalf("mode = %d, want viewStandings", m.mode)
	}
}

func TestUpdate_vKeyOpensResults(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	if m.mode != viewResults {
		t.Fatalf("mode = %d, want viewResults", m.mode)
	}
}

func TestUpdate_bKeyReturnsToSetup(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("b")})
	if m.mode != viewSetup {
		t.Fatalf("mode = %d, want viewSetup after b", m.mode)
	}
}

func TestUpdate_resetClearsResult(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	if m.simulated {
		t.Fatal("simulated should be false after reset")
	}
	if m.result != nil {
		t.Fatal("result should be nil after reset")
	}
}

func TestUpdate_seedChanges(t *testing.T) {
	m := New()
	original := m.seed
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("+")})
	if m.seed != original+1 {
		t.Fatalf("seed = %d, want %d", m.seed, original+1)
	}
}

func TestUpdate_deterministic(t *testing.T) {
	m1 := New()
	m1, _ = m1.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})

	m2 := New()
	m2, _ = m2.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})

	if len(m1.result.Table) != len(m2.result.Table) {
		t.Fatal("non-deterministic table length")
	}
	for i := range m1.result.Table {
		if m1.result.Table[i].Points != m2.result.Table[i].Points {
			t.Fatalf("non-deterministic points at position %d", i)
		}
	}
}

func TestNewWithConfig_usesStorageBackedLeague(t *testing.T) {
	cfg, closeStore, err := app.OpenLocalStore(context.Background(), app.Config{StateDir: t.TempDir()})
	if err != nil {
		t.Fatalf("OpenLocalStore() error = %v", err)
	}
	defer closeStore()

	m := NewWithConfig(cfg)
	if m.clubCount < 3 {
		t.Fatalf("clubCount = %d, want at least 3", m.clubCount)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	if !m.simulated {
		t.Fatal("simulated should be true after storage-backed run")
	}
}

func TestUpdate_hLoadsSavedSeason(t *testing.T) {
	cfg, closeStore, err := app.OpenLocalStore(context.Background(), app.Config{StateDir: t.TempDir()})
	if err != nil {
		t.Fatalf("OpenLocalStore() error = %v", err)
	}
	defer closeStore()

	m := NewWithConfig(cfg)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	if len(m.history) == 0 {
		t.Fatal("expected saved season history after simulate")
	}
	m.simulated = false
	m.result = nil
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if !m.simulated || m.result == nil {
		t.Fatal("expected saved season to load after h")
	}
}

func TestUpdate_cCreatesBranchFromSavedSeason(t *testing.T) {
	cfg, closeStore, err := app.OpenLocalStore(context.Background(), app.Config{StateDir: t.TempDir()})
	if err != nil {
		t.Fatalf("OpenLocalStore() error = %v", err)
	}
	defer closeStore()

	originalBranchID := cfg.State.ActiveBranchID
	m := NewWithConfig(cfg)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	if len(m.history) == 0 {
		t.Fatal("expected saved season history after simulate")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")})
	if cfg.State.ActiveBranchID == originalBranchID {
		t.Fatal("expected shared active branch to change after c")
	}
}

func TestView_setupContainsExpectedContent(t *testing.T) {
	m := New()
	view := m.View(120, 40)
	for _, fragment := range []string{"Season Lab", "Seed", "Simulate"} {
		if !strings.Contains(view, fragment) {
			t.Fatalf("setup view missing %q", fragment)
		}
	}
}

func TestView_standingsContainsTable(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("t")})
	view := m.View(120, 40)
	if !strings.Contains(view, "Season Standings") {
		t.Fatal("standings view missing 'Season Standings'")
	}
}

func TestView_resultsContainsMatchData(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	view := m.View(120, 40)
	if !strings.Contains(view, "Match Results") {
		t.Fatal("results view missing 'Match Results'")
	}
}
