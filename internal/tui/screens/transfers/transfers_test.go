package transfers

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNew_hasData(t *testing.T) {
	m := New()
	if len(m.shortlist) == 0 {
		t.Fatal("shortlist must not be empty")
	}
	if len(m.contracts) == 0 {
		t.Fatal("contracts must not be empty")
	}
}

func TestUpdate_tabCycles(t *testing.T) {
	m := New()
	if m.tab != tabShortlist {
		t.Fatal("tab should start at shortlist")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.tab != tabContracts {
		t.Fatalf("tab = %d, want tabContracts after tab key", m.tab)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.tab != tabShortlist {
		t.Fatalf("tab = %d, want tabShortlist after second tab", m.tab)
	}
}

func TestUpdate_cursorNavigation(t *testing.T) {
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

func TestView_shortlistContainsExpected(t *testing.T) {
	m := New()
	view := m.View(120, 40)
	for _, fragment := range []string{"Transfers", "Shortlist", "OVR"} {
		if !strings.Contains(view, fragment) {
			t.Fatalf("view missing %q", fragment)
		}
	}
}

func TestView_contractsTabContent(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	view := m.View(120, 40)
	if !strings.Contains(view, "Contracts") {
		t.Fatal("contracts tab view missing 'Contracts'")
	}
}
