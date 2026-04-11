package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ismael/football-analytics/internal/app"
	"github.com/ismael/football-analytics/internal/tui/keymap"
	"github.com/ismael/football-analytics/internal/tui/layout"
	"github.com/ismael/football-analytics/internal/tui/screens"
)

type focusArea string

const (
	focusNav  focusArea = "nav"
	focusMain focusArea = "main"
)

// Model is the root Bubble Tea application shell.
type Model struct {
	cfg        app.Config
	sections   []string
	selected   int
	focus      focusArea
	showHelp   bool
	width      int
	height     int
	quitting   bool
	keys       keymap.Map
	statusNote string
}

// NewModel creates the root TUI shell with the default navigation sections.
func NewModel(cfg app.Config) Model {
	return Model{
		cfg:        cfg,
		sections:   append([]string(nil), keymap.DefaultSections...),
		focus:      focusNav,
		keys:       keymap.DefaultMap(),
		statusNote: "Bootstrap shell ready",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch {
		case keymap.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case keymap.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
		case keymap.Matches(msg, m.keys.Back):
			m.focus = focusNav
			m.statusNote = "Focus returned to navigation"
		case keymap.Matches(msg, m.keys.FocusNext):
			if m.focus == focusNav {
				m.focus = focusMain
				m.statusNote = "Main content focused"
			} else {
				m.focus = focusNav
				m.statusNote = "Navigation focused"
			}
		case keymap.Matches(msg, m.keys.Down):
			m.selected = clamp(m.selected+1, 0, len(m.sections)-1)
			m.statusNote = "Section changed"
		case keymap.Matches(msg, m.keys.Up):
			m.selected = clamp(m.selected-1, 0, len(m.sections)-1)
			m.statusNote = "Section changed"
		case keymap.Matches(msg, m.keys.Select):
			m.focus = focusMain
			m.statusNote = "Opened " + m.activeSection()
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	vm := layout.ViewModel{
		AppName:     m.cfg.Name,
		Version:     m.cfg.Version,
		Width:       fallbackSize(m.width, 100),
		Height:      fallbackSize(m.height, 30),
		Sections:    m.sections,
		Selected:    m.selected,
		Active:      m.activeSection(),
		FocusedNav:  m.focus == focusNav,
		MainContent: screens.ContentFor(m.activeSection()),
		StatusNote:  m.statusNote,
		HelpContent: strings.Join(m.keys.ShortHelp(), "  "),
		ShowHelp:    m.showHelp,
		FocusLabel:  string(m.focus),
	}
	return lipgloss.Place(vm.Width, vm.Height, lipgloss.Left, lipgloss.Top, layout.RenderAppShell(vm))
}

func (m Model) activeSection() string {
	if len(m.sections) == 0 {
		return "Dashboard"
	}
	return m.sections[m.selected]
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func fallbackSize(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
