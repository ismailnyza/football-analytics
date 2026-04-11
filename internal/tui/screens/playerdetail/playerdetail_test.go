package playerdetail

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNew_hasPlayers(t *testing.T) {
	m := New()
	if len(m.players) == 0 {
		t.Fatal("demo players must not be empty")
	}
	if m.index != 0 {
		t.Fatal("index should start at 0")
	}
}

func TestUpdate_navigatePlayers(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	if m.index != 1 {
		t.Fatalf("index = %d, want 1 after l", m.index)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if m.index != 0 {
		t.Fatalf("index = %d, want 0 after h", m.index)
	}
}

func TestUpdate_noPastBounds(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("h")})
	if m.index < 0 {
		t.Fatalf("index = %d, must not be negative", m.index)
	}
	// Go to end
	for i := 0; i < 100; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")})
	}
	if m.index >= len(m.players) {
		t.Fatalf("index = %d out of bounds (len=%d)", m.index, len(m.players))
	}
}

func TestView_showsPlayerName(t *testing.T) {
	m := New()
	view := m.View(120, 40)
	if !strings.Contains(view, "Player Detail") {
		t.Fatal("view missing 'Player Detail'")
	}
	if !strings.Contains(view, "Attributes") {
		t.Fatal("view missing 'Attributes'")
	}
}
