package dashboard

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNew_initialState(t *testing.T) {
	m := New()
	if m.summary.WorldName == "" {
		t.Fatal("world name must not be empty")
	}
	if m.scroll != 0 {
		t.Fatal("scroll should start at 0")
	}
}

func TestUpdate_scrollBounds(t *testing.T) {
	m := New()
	// scroll up when at 0 should be no-op
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m.scroll != 0 {
		t.Fatalf("scroll = %d, should stay 0", m.scroll)
	}
	// scroll down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.scroll != 1 {
		t.Fatalf("scroll = %d, want 1", m.scroll)
	}
}

func TestView_containsExpectedContent(t *testing.T) {
	m := New()
	view := m.View(120, 40)
	for _, fragment := range []string{"Dashboard", "World", "Season", "Branch"} {
		if !strings.Contains(view, fragment) {
			t.Fatalf("view missing %q", fragment)
		}
	}
}

func TestRenderProgress_empty(t *testing.T) {
	bar := renderProgress(0, 0, 10)
	if len([]rune(bar)) != 10 {
		t.Fatalf("bar length = %d, want 10", len([]rune(bar)))
	}
}

func TestRenderProgress_full(t *testing.T) {
	bar := renderProgress(10, 10, 10)
	if strings.Contains(bar, "░") {
		t.Fatal("full progress bar should not contain empty segments")
	}
}
