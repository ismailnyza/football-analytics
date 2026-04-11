package club

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNew_hasProfile(t *testing.T) {
	m := New()
	if m.profile.Club.Name == "" {
		t.Fatal("club name must not be empty")
	}
}

func TestUpdate_scrollBounds(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m.scroll != 0 {
		t.Fatal("scroll should not go below 0")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.scroll != 1 {
		t.Fatalf("scroll = %d, want 1", m.scroll)
	}
}

func TestView_containsExpectedSections(t *testing.T) {
	m := New()
	view := m.View(120, 40)
	for _, fragment := range []string{"Club", "Squad Overview", "Recent Form", "Arsenal"} {
		if !strings.Contains(view, fragment) {
			t.Fatalf("view missing %q", fragment)
		}
	}
}
