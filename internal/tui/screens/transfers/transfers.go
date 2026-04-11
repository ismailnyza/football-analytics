// Package transfers provides the Transfers TUI screen — shortlist management,
// player valuation, and contract status overview.
package transfers

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ismael/football-analytics/internal/domain"
	"github.com/ismael/football-analytics/internal/engine/transfer"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117"))
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	valueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	hlStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	greenStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	yellowStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	redStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	helpBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

type tabIndex int

const (
	tabShortlist  tabIndex = 0
	tabContracts  tabIndex = 1
)

var tabLabels = []string{"Shortlist", "Contracts"}

// ShortlistEntry is an entry on the transfer shortlist with computed value.
type ShortlistEntry struct {
	Player domain.Player
	Age    int
	Value  int64
}

// ContractEntry is a player contract for display.
type ContractEntry struct {
	Player   domain.Player
	Contract transfer.Contract
}

// Model is the Bubble Tea model for the Transfers screen.
type Model struct {
	tab       tabIndex
	shortlist []ShortlistEntry
	contracts []ContractEntry
	cursor    int
	scroll    int
}

// New returns a transfers screen with demo data.
func New() Model {
	return Model{
		shortlist: demoShortlist(),
		contracts: demoContracts(),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	kMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch kMsg.String() {
	case "tab":
		m.tab = (m.tab + 1) % tabIndex(len(tabLabels))
		m.cursor = 0
		m.scroll = 0
	case "j", "down":
		listLen := m.currentListLen()
		if m.cursor < listLen-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	}
	return m, nil
}

func (m Model) currentListLen() int {
	switch m.tab {
	case tabContracts:
		return len(m.contracts)
	default:
		return len(m.shortlist)
	}
}

func (m Model) View(width, height int) string {
	// Tab bar
	tabParts := make([]string, len(tabLabels))
	for i, label := range tabLabels {
		if tabIndex(i) == m.tab {
			tabParts[i] = hlStyle.Render(label)
		} else {
			tabParts[i] = dimStyle.Render(label)
		}
	}
	tabBar := "  " + strings.Join(tabParts, "  ")

	var content string
	switch m.tab {
	case tabContracts:
		content = m.renderContracts(height - 6)
	default:
		content = m.renderShortlist(height - 6)
	}

	helpBar := helpBarStyle.Render("tab: switch tab  j/k: navigate")
	return lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Transfers"), "",
		tabBar, "",
		content, "",
		helpBar,
	)
}

func (m Model) renderShortlist(bodyH int) string {
	if len(m.shortlist) == 0 {
		return dimStyle.Render("  Shortlist is empty.")
	}
	header := labelStyle.Render(fmt.Sprintf("  %-20s %-4s %4s %10s", "Name", "Pos", "OVR", "Value"))
	lines := []string{header}
	for i, e := range m.shortlist {
		row := fmt.Sprintf("  %-20s %-4s %4d  £%-9s",
			truncate(e.Player.DisplayName(), 20),
			string(e.Player.PrimaryPosition),
			e.Player.Attributes.Overall,
			formatMoney(e.Value),
		)
		if i == m.cursor {
			lines = append(lines, hlStyle.Render(row))
		} else {
			lines = append(lines, valueStyle.Render(row))
		}
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderContracts(bodyH int) string {
	if len(m.contracts) == 0 {
		return dimStyle.Render("  No contract data.")
	}
	header := labelStyle.Render(fmt.Sprintf("  %-20s %-4s %6s %8s %10s", "Name", "Pos", "Years", "Wage/wk", "Status"))
	lines := []string{header}
	for i, e := range m.contracts {
		status := e.Contract.Status()
		statusStr := string(status)
		row := fmt.Sprintf("  %-20s %-4s %6d  £%-7s %-10s",
			truncate(e.Player.DisplayName(), 20),
			string(e.Player.PrimaryPosition),
			e.Contract.YearsRemaining,
			formatMoney(e.Contract.WeeklyWage),
			statusStr,
		)
		var styled string
		switch status {
		case transfer.StatusSecure:
			styled = greenStyle.Render(row)
		case transfer.StatusMonitor:
			styled = yellowStyle.Render(row)
		default:
			styled = redStyle.Render(row)
		}
		if i == m.cursor {
			styled = hlStyle.Render(row)
		}
		lines = append(lines, styled)
	}
	return strings.Join(lines, "\n")
}

func formatMoney(v int64) string {
	if v >= 1_000_000 {
		return fmt.Sprintf("%.1fm", float64(v)/1_000_000)
	}
	if v >= 1_000 {
		return fmt.Sprintf("%.0fk", float64(v)/1_000)
	}
	return fmt.Sprintf("%d", v)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

func demoShortlist() []ShortlistEntry {
	type spec struct {
		first, last string
		pos         domain.Position
		ovr, pot    int
		age         int
	}
	specs := []spec{
		{"Rodri", "Hernandez", domain.PositionDM, 91, 93, 28},
		{"Erling", "Haaland", domain.PositionST, 92, 94, 24},
		{"Pedri", "Gonzalez", domain.PositionCM, 87, 93, 22},
		{"Gavi", "Paez", domain.PositionCM, 86, 92, 20},
		{"Ruben", "Dias", domain.PositionCB, 88, 89, 27},
	}
	cid := int64(1)
	entries := make([]ShortlistEntry, len(specs))
	for i, s := range specs {
		p := domain.Player{
			ID: int64(i + 1), ClubID: &cid,
			FirstName: s.first, LastName: s.last,
			PrimaryPosition: s.pos,
			Attributes: domain.PlayerAttributes{Overall: s.ovr, Potential: s.pot},
		}
		entries[i] = ShortlistEntry{Player: p, Age: s.age, Value: transfer.MarketValue(p, s.age)}
	}
	return entries
}

func demoContracts() []ContractEntry {
	type spec struct {
		first, last string
		pos         domain.Position
		ovr         int
		years       int
	}
	specs := []spec{
		{"Bukayo", "Saka", domain.PositionRW, 89, 4},
		{"Martin", "Odegaard", domain.PositionAM, 88, 3},
		{"Declan", "Rice", domain.PositionDM, 87, 4},
		{"Gabriel", "Jesus", domain.PositionST, 81, 1},
		{"Jorginho", "Frello", domain.PositionCM, 79, 0},
	}
	cid := int64(1)
	entries := make([]ContractEntry, len(specs))
	for i, s := range specs {
		p := domain.Player{
			ID: int64(i + 1), ClubID: &cid,
			FirstName: s.first, LastName: s.last,
			PrimaryPosition: s.pos,
			Attributes: domain.PlayerAttributes{Overall: s.ovr},
		}
		c := transfer.ContractFor(p, 1, 2024, s.years)
		entries[i] = ContractEntry{Player: p, Contract: c}
	}
	return entries
}
