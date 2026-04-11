package world

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNew_hasLeagues(t *testing.T) {
	m := New()
	if len(m.leagues) == 0 {
		t.Fatal("leagues must not be empty")
	}
}

func TestUpdate_scrollBounds(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m.scroll < 0 {
		t.Fatal("scroll must not go below 0")
	}
}

func TestView_containsExpectedContent(t *testing.T) {
	m := New()
	view := m.View(120, 40)
	for _, fragment := range []string{"World", "Competitions", "Premier League"} {
		if !strings.Contains(view, fragment) {
			t.Fatalf("view missing %q", fragment)
		}
	}
}
