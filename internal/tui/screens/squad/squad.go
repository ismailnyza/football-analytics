// Package squad provides the Squad TUI screen showing the active club's
// player roster with sortable columns and selection.
package squad

import (
	"fmt"
	"slices"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ismael/football-analytics/internal/domain"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117"))
	headerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Bold(true)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	helpBarStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

type sortField int

const (
	sortByOvr  sortField = 0
	sortByPos  sortField = 1
	sortByName sortField = 2
)

var sortLabels = []string{"OVR", "POS", "Name"}

// Model is the Bubble Tea model for the Squad screen.
type Model struct {
	players  []domain.Player
	sorted   []domain.Player
	cursor   int
	scroll   int
	sortBy   sortField
	pageSize int
}

// New returns a squad screen pre-loaded with a demo roster.
func New() Model {
	return NewWithPlayers(demoSquad())
}

// NewWithPlayers returns a squad screen backed by caller-provided players.
func NewWithPlayers(players []domain.Player) Model {
	m := Model{
		players:  append([]domain.Player(nil), players...),
		sortBy:   sortByOvr,
		pageSize: 20,
	}
	m.sorted = sortPlayers(m.players, sortByOvr)
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	kMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch kMsg.String() {
	case "j", "down":
		if m.cursor < len(m.sorted)-1 {
			m.cursor++
			if m.cursor >= m.scroll+m.pageSize {
				m.scroll++
			}
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
			if m.cursor < m.scroll {
				m.scroll--
			}
		}
	case "s":
		m.sortBy = (m.sortBy + 1) % sortField(len(sortLabels))
		m.sorted = sortPlayers(m.players, m.sortBy)
		m.cursor = 0
		m.scroll = 0
	}
	return m, nil
}

func (m Model) View(width, height int) string {
	pageSize := height - 7
	if pageSize < 4 {
		pageSize = 4
	}
	m.pageSize = pageSize

	header := fmt.Sprintf("  %-20s %-4s %4s %4s %4s %4s %4s %4s",
		"Name", "Pos", "OVR", "PAC", "SHO", "PAS", "DEF", "KEP")

	lines := []string{
		titleStyle.Render("Squad"),
		"",
		headerStyle.Render(header),
	}

	end := m.scroll + pageSize
	if end > len(m.sorted) {
		end = len(m.sorted)
	}
	for i, p := range m.sorted[m.scroll:end] {
		idx := m.scroll + i
		row := fmt.Sprintf("  %-20s %-4s %4d %4d %4d %4d %4d %4d",
			truncate(p.DisplayName(), 20),
			string(p.PrimaryPosition),
			p.Attributes.Overall,
			p.Attributes.Pace,
			p.Attributes.Shooting,
			p.Attributes.Passing,
			p.Attributes.Defending,
			p.Attributes.Keeping,
		)
		if idx == m.cursor {
			lines = append(lines, selectedStyle.Render(row))
		} else {
			lines = append(lines, row)
		}
	}

	if len(m.sorted) > pageSize {
		lines = append(lines, "", dimStyle.Render(fmt.Sprintf(
			"  %d–%d of %d players  (j/k to scroll  s: cycle sort)",
			m.scroll+1, end, len(m.sorted),
		)))
	}

	sortHint := "Sort: " + sortLabels[m.sortBy]
	helpBar := helpBarStyle.Render("j/k: navigate  s: cycle sort [" + sortHint + "]")
	return lipgloss.JoinVertical(lipgloss.Left, strings.Join(lines, "\n"), "", helpBar)
}

func sortPlayers(players []domain.Player, by sortField) []domain.Player {
	out := make([]domain.Player, len(players))
	copy(out, players)
	slices.SortFunc(out, func(a, b domain.Player) int {
		switch by {
		case sortByPos:
			if string(a.PrimaryPosition) != string(b.PrimaryPosition) {
				if string(a.PrimaryPosition) < string(b.PrimaryPosition) {
					return -1
				}
				return 1
			}
		case sortByName:
			na, nb := a.DisplayName(), b.DisplayName()
			if na < nb {
				return -1
			}
			if na > nb {
				return 1
			}
			return 0
		}
		// Default: descending OVR
		return b.Attributes.Overall - a.Attributes.Overall
	})
	return out
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func demoSquad() []domain.Player {
	type spec struct {
		first, last string
		pos         domain.Position
		ovr         int
	}
	specs := []spec{
		{"David", "Raya", domain.PositionGK, 83},
		{"Aaron", "Ramsdale", domain.PositionGK, 78},
		{"Ben", "White", domain.PositionRB, 82},
		{"William", "Saliba", domain.PositionCB, 88},
		{"Gabriel", "Magalhaes", domain.PositionCB, 85},
		{"Oleksandr", "Zinchenko", domain.PositionLB, 80},
		{"Kieran", "Tierney", domain.PositionLB, 75},
		{"Thomas", "Partey", domain.PositionDM, 82},
		{"Jorginho", "Frello", domain.PositionCM, 79},
		{"Martin", "Odegaard", domain.PositionAM, 88},
		{"Fabio", "Vieira", domain.PositionAM, 76},
		{"Bukayo", "Saka", domain.PositionRW, 89},
		{"Leandro", "Trossard", domain.PositionLW, 83},
		{"Gabriel", "Martinelli", domain.PositionLW, 85},
		{"Kai", "Havertz", domain.PositionST, 82},
		{"Eddie", "Nketiah", domain.PositionST, 78},
		{"Gabriel", "Jesus", domain.PositionST, 81},
		{"Declan", "Rice", domain.PositionDM, 87},
		{"Jurrien", "Timber", domain.PositionRB, 83},
		{"Takehiro", "Tomiyasu", domain.PositionCB, 79},
	}
	players := make([]domain.Player, len(specs))
	cid := int64(1)
	for i, s := range specs {
		players[i] = domain.Player{
			ID:              int64(i + 1),
			ClubID:          &cid,
			FirstName:       s.first,
			LastName:        s.last,
			PrimaryPosition: s.pos,
			Attributes: domain.PlayerAttributes{
				Overall:   s.ovr,
				Pace:      clamp(s.ovr-2+i%5, 60, 99),
				Shooting:  clamp(s.ovr-3+i%7, 50, 99),
				Passing:   clamp(s.ovr-2+i%6, 55, 99),
				Defending: clamp(s.ovr-4+i%8, 45, 99),
				Keeping:   gkStat(s.pos, s.ovr),
			},
		}
	}
	return players
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func gkStat(pos domain.Position, ovr int) int {
	if pos == domain.PositionGK {
		return ovr + 5
	}
	return 40
}
