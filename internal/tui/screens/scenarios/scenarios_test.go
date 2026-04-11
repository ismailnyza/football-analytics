package scenarios

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ismael/football-analytics/internal/app"
	"github.com/ismael/football-analytics/internal/domain"
)

func TestNew_hasBranches(t *testing.T) {
	m := New()
	if len(m.branches) == 0 {
		t.Fatal("must have at least one branch")
	}
	if m.active == 0 {
		t.Fatal("active branch must be set")
	}
}

func TestUpdate_navigation(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", m.cursor)
	}
}

func TestUpdate_activateBranch(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.active != m.branches[1].ID {
		t.Fatalf("active = %d, want %d", m.active, m.branches[1].ID)
	}
}

func TestView_containsExpectedContent(t *testing.T) {
	m := New()
	view := m.View(120, 40)
	for _, fragment := range []string{"Scenarios", "Branches", "main"} {
		if !strings.Contains(view, fragment) {
			t.Fatalf("view missing %q", fragment)
		}
	}
}

func TestNewWithConfig_activateBranchUpdatesSharedState(t *testing.T) {
	cfg, closeStore, err := app.OpenLocalStore(context.Background(), app.Config{StateDir: t.TempDir()})
	if err != nil {
		t.Fatalf("OpenLocalStore() error = %v", err)
	}
	defer closeStore()

	branch, err := cfg.Repo.CreateBranch(context.Background(), domain.Branch{
		WorldID: cfg.CurrentWorldID(),
		Name:    "alt",
	})
	if err != nil {
		t.Fatalf("CreateBranch() error = %v", err)
	}

	m := NewWithConfig(cfg)
	for i, entry := range m.branches {
		if entry.ID == branch.ID {
			m.cursor = i
			break
		}
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cfg.State.ActiveBranchID != branch.ID {
		t.Fatalf("shared ActiveBranchID = %d, want %d", cfg.State.ActiveBranchID, branch.ID)
	}
}
