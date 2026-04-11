package keymap

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Map defines the root app shell bindings.
type Map struct {
	Up        key.Binding
	Down      key.Binding
	FocusNext key.Binding
	Select    key.Binding
	Back      key.Binding
	Help      key.Binding
	Quit      key.Binding
}

// DefaultMap returns the shell bindings used by the root app.
func DefaultMap() Map {
	return Map{
		Up:        key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("k/up", "move up")),
		Down:      key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("j/down", "move down")),
		FocusNext: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "toggle focus")),
		Select:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open section")),
		Back:      key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "focus nav")),
		Help:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
		Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	}
}

// Matches reports whether a key message satisfies a binding.
func Matches(msg tea.KeyMsg, binding key.Binding) bool {
	return key.Matches(msg, binding)
}

// ShortHelp returns the compact help labels for the active shell bindings.
func (m Map) ShortHelp() []string {
	bindings := []key.Binding{m.Up, m.Down, m.FocusNext, m.Select, m.Back, m.Help, m.Quit}
	items := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		desc := binding.Help()
		items = append(items, desc.Key+": "+desc.Desc)
	}
	return items
}

// SearchableSections returns sections matching a loose user query.
func SearchableSections(query string) []string {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return append([]string(nil), DefaultSections...)
	}
	filtered := make([]string, 0, len(DefaultSections))
	for _, section := range DefaultSections {
		if strings.Contains(strings.ToLower(section), query) {
			filtered = append(filtered, section)
		}
	}
	return filtered
}
