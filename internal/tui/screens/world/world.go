// Package world provides the World TUI screen — global simulation state,
// competition overview, and branch management summary.
package world

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117"))
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	valueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	accentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

// LeagueSnapshot shows season standing for a competition.
type LeagueSnapshot struct {
	Name   string
	Leader string
	Points int
	Season string
}

// Model is the Bubble Tea model for the World screen.
type Model struct {
	worldName   string
	branches    int
	currentYear int
	leagues     []LeagueSnapshot
	scroll      int
}

// New returns a world screen with demo data.
func New() Model {
	return NewWithData(
		"Demo World",
		1,
		2024,
		[]LeagueSnapshot{
			{"Premier League", "Arsenal FC", 72, "2024/25"},
			{"La Liga", "Real Madrid", 68, "2024/25"},
			{"Bundesliga", "Bayern Munich", 65, "2024/25"},
			{"Serie A", "Inter Milan", 61, "2024/25"},
			{"Ligue 1", "Paris SG", 70, "2024/25"},
		},
	)
}

// NewWithData returns a world screen with caller-provided summary data.
func NewWithData(worldName string, branches int, currentYear int, leagues []LeagueSnapshot) Model {
	return Model{
		worldName:   worldName,
		branches:    branches,
		currentYear: currentYear,
		leagues:     append([]LeagueSnapshot(nil), leagues...),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	kMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch kMsg.String() {
	case "j", "down":
		m.scroll++
	case "k", "up":
		if m.scroll > 0 {
			m.scroll--
		}
	}
	return m, nil
}

func (m Model) View(width, height int) string {
	lines := []string{
		titleStyle.Render("World"),
		"",
		labelStyle.Render("World Name"),
		accentStyle.Render("  " + m.worldName),
		"",
		labelStyle.Render("Simulation Year"),
		valueStyle.Render(fmt.Sprintf("  %d", m.currentYear)),
		"",
		labelStyle.Render("Branches"),
		valueStyle.Render(fmt.Sprintf("  %d active", m.branches)),
		"",
		titleStyle.Render("Competitions"),
	}

	for _, l := range m.leagues {
		lines = append(lines,
			fmt.Sprintf("  %-20s  Leader: %-18s  %3d pts  (%s)",
				l.Name, l.Leader, l.Points, l.Season),
		)
	}

	lines = append(lines, "",
		dimStyle.Render("  Navigate to Scenarios to create or restore branches."),
	)

	bodyH := height - 4
	if bodyH < 4 {
		bodyH = 4
	}
	total := len(lines)
	maxScroll := total - bodyH
	if maxScroll < 0 {
		maxScroll = 0
	}
	start := m.scroll
	if start > maxScroll {
		start = maxScroll
	}
	end := start + bodyH
	if end > total {
		end = total
	}

	body := strings.Join(lines[start:end], "\n")
	helpBar := helpBarStyle.Render("j/k: scroll")
	return lipgloss.JoinVertical(lipgloss.Left, body, "", helpBar)
}
