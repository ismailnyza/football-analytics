package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ismael/football-analytics/internal/app"
)

func TestModelNavigation(t *testing.T) {
	model := NewModel(app.DefaultConfig())

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	updated := next.(Model)
	if updated.selected != 1 {
		t.Fatalf("selected = %d, want 1", updated.selected)
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated = next.(Model)
	if updated.focus != focusMain {
		t.Fatalf("focus = %q, want %q", updated.focus, focusMain)
	}

	next, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	updated = next.(Model)
	if !updated.showHelp {
		t.Fatal("showHelp = false, want true")
	}
}

func TestViewIncludesShellRegions(t *testing.T) {
	model := NewModel(app.DefaultConfig())
	model.width = 100
	model.height = 30

	view := model.View()
	required := []string{
		"Football Simulation Engine",
		"Navigation",
		"Main Content",
		"Status",
		"Dashboard",
		"Demo World",
	}
	for _, fragment := range required {
		if !strings.Contains(view, fragment) {
			t.Fatalf("View() missing %q", fragment)
		}
	}
}
