// Package playerdetail provides the Player Detail TUI screen showing a single
// player's identity, attributes, and form indicators.
package playerdetail

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ismael/football-analytics/internal/domain"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117"))
	labelStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Width(14)
	valueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	accentStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
	barFillStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("62"))
	barEmptyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("237"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpBarStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

// Model is the Bubble Tea model for the Player Detail screen.
type Model struct {
	players []domain.Player
	index   int
	scroll  int
}

// New returns a player detail screen pre-loaded with demo players.
func New() Model {
	return NewWithPlayers(demoPlayers())
}

// NewWithPlayers returns a player detail screen backed by caller-provided players.
func NewWithPlayers(players []domain.Player) Model {
	return Model{players: append([]domain.Player(nil), players...)}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	kMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch kMsg.String() {
	case "l", "right":
		if m.index < len(m.players)-1 {
			m.index++
			m.scroll = 0
		}
	case "h", "left":
		if m.index > 0 {
			m.index--
			m.scroll = 0
		}
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
	if len(m.players) == 0 {
		return dimStyle.Render("No players loaded.")
	}

	p := m.players[m.index]
	navHint := fmt.Sprintf("Player %d / %d", m.index+1, len(m.players))

	lines := []string{
		titleStyle.Render("Player Detail"),
		"",
		accentStyle.Render(p.DisplayName()),
		dimStyle.Render(navHint),
		"",
		row("Position", string(p.PrimaryPosition)),
		row("Overall", fmt.Sprintf("%d", p.Attributes.Overall)),
		row("Potential", fmt.Sprintf("%d", p.Attributes.Potential)),
		"",
		titleStyle.Render("Attributes"),
		attrBar("Pace     ", p.Attributes.Pace),
		attrBar("Shooting ", p.Attributes.Shooting),
		attrBar("Passing  ", p.Attributes.Passing),
		attrBar("Defending", p.Attributes.Defending),
		attrBar("Keeping  ", p.Attributes.Keeping),
		"",
		dimStyle.Render("← / → to browse players"),
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
	helpBar := helpBarStyle.Render("h/l: prev/next player  j/k: scroll")
	return lipgloss.JoinVertical(lipgloss.Left, body, "", helpBar)
}

func row(label, val string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("  "+label),
		valueStyle.Render(val),
	)
}

func attrBar(label string, value int) string {
	const barWidth = 20
	filled := value * barWidth / 100
	if filled > barWidth {
		filled = barWidth
	}
	bar := barFillStyle.Render(strings.Repeat("█", filled)) +
		barEmptyStyle.Render(strings.Repeat("░", barWidth-filled))
	return fmt.Sprintf("  %-10s %s  %d", label, bar, value)
}

func demoPlayers() []domain.Player {
	type spec struct {
		first, last             string
		pos                     domain.Position
		ovr, pot                int
		pac, sho, pas, def, kep int
	}
	specs := []spec{
		{"Bukayo", "Saka", domain.PositionRW, 89, 93, 88, 82, 85, 60, 40},
		{"Martin", "Odegaard", domain.PositionAM, 88, 90, 78, 83, 91, 55, 40},
		{"Declan", "Rice", domain.PositionDM, 87, 89, 80, 72, 84, 82, 40},
		{"William", "Saliba", domain.PositionCB, 88, 91, 76, 48, 71, 88, 40},
		{"David", "Raya", domain.PositionGK, 83, 85, 55, 45, 65, 72, 87},
	}
	cid := int64(1)
	players := make([]domain.Player, len(specs))
	for i, s := range specs {
		players[i] = domain.Player{
			ID: int64(i + 1), ClubID: &cid,
			FirstName: s.first, LastName: s.last,
			PrimaryPosition: s.pos,
			Attributes: domain.PlayerAttributes{
				Overall: s.ovr, Potential: s.pot,
				Pace: s.pac, Shooting: s.sho, Passing: s.pas,
				Defending: s.def, Keeping: s.kep,
			},
		}
	}
	return players
}
