package squad

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNew_hasPlayers(t *testing.T) {
	m := New()
	if len(m.players) == 0 {
		t.Fatal("demo squad must not be empty")
	}
	if len(m.sorted) == 0 {
		t.Fatal("sorted list must not be empty")
	}
}

func TestUpdate_cursorNavigation(t *testing.T) {
	m := New()
	if m.cursor != 0 {
		t.Fatal("cursor should start at 0")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	if m.cursor != 1 {
		t.Fatalf("cursor = %d, want 1 after j", m.cursor)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m.cursor != 0 {
		t.Fatalf("cursor = %d, want 0 after k", m.cursor)
	}
}

func TestUpdate_cursorNoNegative(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	if m.cursor < 0 {
		t.Fatalf("cursor = %d, must not be negative", m.cursor)
	}
}

func TestUpdate_sortCycles(t *testing.T) {
	m := New()
	initial := m.sortBy
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	if m.sortBy == initial {
		t.Fatal("sort field should change after s")
	}
	// Cycle through all sort modes back to initial.
	for i := 0; i < len(sortLabels)-1; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("s")})
	}
	if m.sortBy != initial {
		t.Fatalf("sortBy = %d, want %d after full cycle", m.sortBy, initial)
	}
}

func TestUpdate_sortByOvrDescending(t *testing.T) {
	m := New()
	// Default sort is by OVR descending.
	for i := 1; i < len(m.sorted); i++ {
		if m.sorted[i].Attributes.Overall > m.sorted[i-1].Attributes.Overall {
			t.Fatalf("OVR sort not descending at index %d: %d > %d",
				i, m.sorted[i].Attributes.Overall, m.sorted[i-1].Attributes.Overall)
		}
	}
}

func TestView_containsHeaderAndPlayers(t *testing.T) {
	m := New()
	view := m.View(120, 40)
	for _, fragment := range []string{"Squad", "Name", "OVR"} {
		if !strings.Contains(view, fragment) {
			t.Fatalf("view missing %q", fragment)
		}
	}
}
