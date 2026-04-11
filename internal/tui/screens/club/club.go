// Package club provides the Club TUI screen showing a club's profile,
// squad summary, recent results, and key indicators.
package club

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ismael/football-analytics/internal/domain"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117"))
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Width(16)
	valueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	accentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	resultW      = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
	resultD      = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	resultL      = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

// RecentResult holds a single historical match result for display.
type RecentResult struct {
	Opponent  string
	HomeGoals int
	AwayGoals int
	Home      bool // true = this club was home
}

// ClubProfile holds the display data for a club.
type ClubProfile struct {
	Club       domain.Club
	SquadSize  int
	AverageOvr int
	Budget     int64
	RecentForm []RecentResult
}

// Model is the Bubble Tea model for the Club screen.
type Model struct {
	profile ClubProfile
	scroll  int
}

// New returns a club screen pre-loaded with demo data.
func New() Model {
	return NewWithProfile(demoProfile())
}

// NewWithProfile returns a club screen backed by caller-provided data.
func NewWithProfile(profile ClubProfile) Model {
	return Model{profile: profile}
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
	p := m.profile
	c := p.Club

	lines := []string{
		titleStyle.Render("Club"),
		"",
		accentStyle.Render(c.Name),
		labelStyle.Render("  Short Name") + valueStyle.Render(c.ShortName),
		labelStyle.Render("  Country") + valueStyle.Render(c.Country),
		labelStyle.Render("  Division") + valueStyle.Render(c.Division),
		"",
		titleStyle.Render("Squad Overview"),
		labelStyle.Render("  Squad Size") + valueStyle.Render(fmt.Sprintf("%d players", p.SquadSize)),
		labelStyle.Render("  Average OVR") + valueStyle.Render(fmt.Sprintf("%d", p.AverageOvr)),
		labelStyle.Render("  Budget") + accentStyle.Render(fmt.Sprintf("£%dm", p.Budget/1_000_000)),
		"",
		titleStyle.Render("Recent Form"),
	}

	if len(p.RecentForm) == 0 {
		lines = append(lines, dimStyle.Render("  No recent results."))
	} else {
		for _, r := range p.RecentForm {
			var homeGoals, awayGoals int
			if r.Home {
				homeGoals, awayGoals = r.HomeGoals, r.AwayGoals
			} else {
				homeGoals, awayGoals = r.AwayGoals, r.HomeGoals
			}
			venue := "H"
			if !r.Home {
				venue = "A"
			}
			score := fmt.Sprintf("%d–%d", homeGoals, awayGoals)
			entry := fmt.Sprintf("  [%s]  vs %-12s  %s", venue, r.Opponent, score)
			switch {
			case homeGoals > awayGoals:
				lines = append(lines, resultW.Render(entry+" W"))
			case homeGoals < awayGoals:
				lines = append(lines, resultL.Render(entry+" L"))
			default:
				lines = append(lines, resultD.Render(entry+" D"))
			}
		}
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

func demoProfile() ClubProfile {
	return ClubProfile{
		Club: domain.Club{
			ID: 1, BranchID: 1,
			Name: "Arsenal FC", ShortName: "ARS",
			Country: "England", Division: "Premier League",
		},
		SquadSize:  20,
		AverageOvr: 83,
		Budget:     150_000_000,
		RecentForm: []RecentResult{
			{"Liverpool FC", 2, 1, true},
			{"Manchester City", 1, 1, false},
			{"Chelsea FC", 3, 0, true},
			{"Tottenham", 2, 2, false},
			{"Newcastle Utd", 1, 0, true},
		},
	}
}
