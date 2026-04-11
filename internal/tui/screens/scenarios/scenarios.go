// Package scenarios provides the Scenarios TUI screen — branch management,
// checkpoint creation, and timeline navigation.
package scenarios

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ismael/football-analytics/internal/app"
	"github.com/ismael/football-analytics/internal/domain"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117"))
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	valueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	hlStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("230")).Background(lipgloss.Color("62"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	activeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true)
	helpBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
)

// BranchEntry represents a simulation branch/scenario.
type BranchEntry struct {
	ID       int64
	Name     string
	IsRoot   bool
	Year     int
	ParentID *int64
}

// Model is the Bubble Tea model for the Scenarios screen.
type Model struct {
	cfg      app.Config
	branches []BranchEntry
	cursor   int
	active   int64 // active branch ID
}

// New returns a scenarios screen with demo data.
func New() Model {
	branches := []BranchEntry{
		{ID: 1, Name: "main", IsRoot: true, Year: 2024},
		{ID: 2, Name: "no-haaland-transfer", Year: 2023, ParentID: ptr(int64(1))},
		{ID: 3, Name: "double-win-2023", Year: 2023, ParentID: ptr(int64(1))},
	}
	return Model{branches: branches, active: 1}
}

// NewWithConfig returns a scenarios screen backed by repository branches when available.
func NewWithConfig(cfg app.Config) Model {
	m := New()
	m.cfg = cfg
	if cfg.Repo == nil || cfg.CurrentWorldID() == 0 {
		return m
	}
	branches, err := cfg.Repo.ListBranches(context.Background(), cfg.CurrentWorldID())
	if err != nil || len(branches) == 0 {
		return m
	}
	m.branches = make([]BranchEntry, 0, len(branches))
	for _, branch := range branches {
		m.branches = append(m.branches, branchEntry(branch))
	}
	m.active = cfg.CurrentBranchID()
	for i, branch := range m.branches {
		if branch.ID == m.active {
			m.cursor = i
			break
		}
	}
	return m
}

func ptr(v int64) *int64 { return &v }

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	kMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch kMsg.String() {
	case "j", "down":
		if m.cursor < len(m.branches)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "enter":
		if m.cursor < len(m.branches) {
			m.active = m.branches[m.cursor].ID
			if m.cfg.State != nil {
				m.cfg.State.ActiveBranchID = m.branches[m.cursor].ID
				m.cfg.State.ActiveBranchName = m.branches[m.cursor].Name
			}
		}
	}
	return m, nil
}

func branchEntry(branch domain.Branch) BranchEntry {
	return BranchEntry{
		ID:       branch.ID,
		Name:     branch.Name,
		IsRoot:   branch.IsRoot(),
		Year:     branch.CreatedAt.Year(),
		ParentID: branch.ParentBranchID,
	}
}

func (m Model) View(width, height int) string {
	lines := []string{
		titleStyle.Render("Scenarios & Branches"),
		"",
		labelStyle.Render("Active Branch"),
		activeStyle.Render("  " + m.activeName()),
		"",
		titleStyle.Render("All Branches"),
	}

	for i, b := range m.branches {
		indicator := "  "
		if b.ID == m.active {
			indicator = "* "
		}
		parentInfo := ""
		if b.ParentID != nil {
			parentInfo = fmt.Sprintf(" (from branch #%d)", *b.ParentID)
		}
		label := fmt.Sprintf("%s%-30s  year %d%s", indicator, b.Name, b.Year, parentInfo)
		if b.IsRoot {
			label += "  [root]"
		}
		if i == m.cursor {
			lines = append(lines, hlStyle.Render("  "+label))
		} else if b.ID == m.active {
			lines = append(lines, activeStyle.Render("  "+label))
		} else {
			lines = append(lines, valueStyle.Render("  "+label))
		}
	}

	lines = append(lines, "",
		dimStyle.Render("  enter: activate branch"),
	)

	helpBar := helpBarStyle.Render("j/k: navigate  enter: activate branch")
	return lipgloss.JoinVertical(lipgloss.Left,
		strings.Join(lines, "\n"), "", helpBar)
}

func (m Model) activeName() string {
	for _, b := range m.branches {
		if b.ID == m.active {
			return b.Name
		}
	}
	return "unknown"
}
