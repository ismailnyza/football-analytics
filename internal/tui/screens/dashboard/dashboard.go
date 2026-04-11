// Package dashboard provides the Dashboard TUI screen — an overview of the
// active world, current season status, and navigation shortcuts.
package dashboard

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
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	accentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	helpBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

// WorldSummary holds the high-level world state shown on the dashboard.
type WorldSummary struct {
	WorldName     string
	Branch        string
	Season        string
	MatchesPlayed int
	MatchesTotal  int
	LeaderName    string
	LeaderPoints  int
	Clubs         int
}

// Model is the Bubble Tea model for the Dashboard screen.
type Model struct {
	summary WorldSummary
	scroll  int
	width   int
	height  int
}

// New returns a dashboard with demo world state.
func New() Model {
	return NewWithSummary(WorldSummary{
		WorldName:     "Demo World",
		Branch:        "main",
		Season:        "2024/25",
		MatchesPlayed: 0,
		MatchesTotal:  30,
		LeaderName:    "—",
		LeaderPoints:  0,
		Clubs:         6,
	})
}

// NewWithSummary returns a dashboard with caller-provided world state.
func NewWithSummary(summary WorldSummary) Model {
	return Model{summary: summary}
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
	m.width = width
	m.height = height

	s := m.summary
	progressBar := renderProgress(s.MatchesPlayed, s.MatchesTotal, 30)

	lines := []string{
		titleStyle.Render("Dashboard"),
		"",
		labelStyle.Render("World"),
		valueStyle.Render("  " + s.WorldName),
		"",
		labelStyle.Render("Branch"),
		valueStyle.Render("  " + s.Branch),
		"",
		labelStyle.Render("Season"),
		accentStyle.Render("  " + s.Season),
		"",
		labelStyle.Render("Season Progress"),
		fmt.Sprintf("  %s  %d / %d matches", progressBar, s.MatchesPlayed, s.MatchesTotal),
		"",
		labelStyle.Render("Clubs in League"),
		valueStyle.Render(fmt.Sprintf("  %d", s.Clubs)),
		"",
		labelStyle.Render("Current Leader"),
		accentStyle.Render(fmt.Sprintf("  %s  (%d pts)", s.LeaderName, s.LeaderPoints)),
		"",
		dimStyle.Render("  Navigate to Match Lab or Season Lab to run simulations."),
		dimStyle.Render("  Use Squad, Player Detail, and Club screens to manage your world."),
	}

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

func renderProgress(done, total, width int) string {
	if total <= 0 {
		return strings.Repeat("░", width)
	}
	filled := done * width / total
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}
