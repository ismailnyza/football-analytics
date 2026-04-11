// Package matchlab provides the Match Lab interactive screen for the TUI.
// It lets the user configure two demo teams, pick a formation, set a seed,
// run a deterministic simulation, and browse the resulting score, stats, and
// event log — all without leaving the terminal.
package matchlab

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ismael/football-analytics/internal/domain"
	"github.com/ismael/football-analytics/internal/engine/match"
)

// ---------------------------------------------------------------------------
// Styles
// ---------------------------------------------------------------------------

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("117"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	scoreStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("220"))

	statLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Width(16)

	statValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Width(6).
			Align(lipgloss.Right)

	eventGoalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Bold(true)

	eventCardStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

	eventInjuryStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("208"))

	eventDefaultStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))

	paneStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2)

	activePaneStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	helpBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))
)

// ---------------------------------------------------------------------------
// Pane identifiers
// ---------------------------------------------------------------------------

type pane int

const (
	paneSetup  pane = 0
	paneResult pane = 1
)

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

// Model is the Bubble Tea model for the Match Lab screen.
type Model struct {
	activePaneFocus pane
	formations      []string
	formationIdx    int
	seed            int64
	simulated       bool
	summary         *match.Summary
	homeClub        domain.Club
	awayClub        domain.Club
	homeSquad       []domain.Player
	awaySquad       []domain.Player
	eventScroll     int
	maxEventLines   int
	width           int
	height          int
}

// New returns an initialised Match Lab model with two demo squads pre-loaded.
func New() Model {
	home, homeSquad := demoTeam(1, "Arsenal FC", "ARS")
	away, awaySquad := demoTeam(2, "Liverpool FC", "LIV")
	return Model{
		activePaneFocus: paneSetup,
		formations:      []string{"4-3-3", "4-2-3-1", "4-4-2"},
		formationIdx:    0,
		seed:            42,
		homeClub:        home,
		awayClub:        away,
		homeSquad:       homeSquad,
		awaySquad:       awaySquad,
		maxEventLines:   18,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if m.activePaneFocus == paneSetup {
				m.activePaneFocus = paneResult
			} else {
				m.activePaneFocus = paneSetup
			}
		case "left", "h":
			if m.activePaneFocus == paneSetup {
				m.formationIdx = (m.formationIdx - 1 + len(m.formations)) % len(m.formations)
			}
		case "right", "l":
			if m.activePaneFocus == paneSetup {
				m.formationIdx = (m.formationIdx + 1) % len(m.formations)
			}
		case "up", "k":
			if m.activePaneFocus == paneResult && m.eventScroll > 0 {
				m.eventScroll--
			}
		case "down", "j":
			if m.activePaneFocus == paneResult && m.simulated {
				maxScroll := len(m.summary.Events) - m.maxEventLines
				if maxScroll < 0 {
					maxScroll = 0
				}
				if m.eventScroll < maxScroll {
					m.eventScroll++
				}
			}
		case "s":
			m = m.runSimulation()
		case "r":
			m.simulated = false
			m.summary = nil
			m.eventScroll = 0
		case "+", "=":
			m.seed++
		case "-":
			if m.seed > 0 {
				m.seed--
			}
		}
	}
	return m, nil
}

func (m Model) runSimulation() Model {
	formation := m.formations[m.formationIdx]
	homeLineup, err := match.SelectLineup(m.homeSquad, formation, nil)
	if err != nil {
		return m
	}
	awayLineup, err := match.SelectLineup(m.awaySquad, formation, nil)
	if err != nil {
		return m
	}
	homePlan := match.BuildTeamPlan(m.homeClub, homeLineup)
	awayPlan := match.BuildTeamPlan(m.awayClub, awayLineup)
	summary := match.RunTickLoop(m.seed, homePlan, awayPlan, 0)
	m.summary = &summary
	m.simulated = true
	m.eventScroll = 0
	return m
}

// View renders the full Match Lab panel.
func (m Model) View(width, height int) string {
	m.width = width
	m.height = height

	availW := width - 4
	if availW < 50 {
		availW = 50
	}
	setupW := availW * 2 / 5
	resultW := availW - setupW - 2

	bodyH := height - 5
	if bodyH < 12 {
		bodyH = 12
	}
	m.maxEventLines = bodyH - 14
	if m.maxEventLines < 4 {
		m.maxEventLines = 4
	}

	setupContent := m.renderSetup(setupW - 6)
	resultContent := m.renderResult(resultW-6, bodyH-4)

	var setupFrame, resultFrame lipgloss.Style
	if m.activePaneFocus == paneSetup {
		setupFrame = activePaneStyle
		resultFrame = paneStyle
	} else {
		setupFrame = paneStyle
		resultFrame = activePaneStyle
	}

	leftPanel := setupFrame.Width(setupW).Height(bodyH).Render(setupContent)
	rightPanel := resultFrame.Width(resultW).Height(bodyH).Render(resultContent)

	body := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel)

	helpBar := helpBarStyle.Render(
		"tab: switch pane  s: simulate  r: reset  ←/→: formation  +/-: seed  j/k: scroll events",
	)

	return lipgloss.JoinVertical(lipgloss.Left, body, "", helpBar)
}

// ---------------------------------------------------------------------------
// Pane renderers
// ---------------------------------------------------------------------------

func (m Model) renderSetup(contentW int) string {
	lines := []string{
		titleStyle.Render("Match Lab Setup"),
		"",
		labelStyle.Render("Home Team"),
		valueStyle.Render("  " + m.homeClub.Name),
		"",
		labelStyle.Render("Away Team"),
		valueStyle.Render("  " + m.awayClub.Name),
		"",
		labelStyle.Render("Formation"),
		m.renderFormationSelector(),
		"",
		labelStyle.Render("Seed"),
		valueStyle.Render(fmt.Sprintf("  %d", m.seed)),
		"",
		dimStyle.Render("(+/- to change seed)"),
		"",
		"",
		highlightStyle.Render(" [s] Simulate "),
	}
	if m.simulated {
		lines = append(lines, "", dimStyle.Render(" [r] Reset results"))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderFormationSelector() string {
	parts := make([]string, len(m.formations))
	for i, f := range m.formations {
		if i == m.formationIdx {
			parts[i] = highlightStyle.Render(f)
		} else {
			parts[i] = dimStyle.Render(f)
		}
	}
	return "  " + strings.Join(parts, "  ")
}

func (m Model) renderResult(contentW, contentH int) string {
	if !m.simulated || m.summary == nil {
		return strings.Join([]string{
			titleStyle.Render("Match Result"),
			"",
			dimStyle.Render("No match simulated yet."),
			"",
			dimStyle.Render("Configure teams on the left"),
			dimStyle.Render("and press [s] to simulate."),
		}, "\n")
	}

	s := m.summary
	homeShort := m.homeClub.ShortName
	awayShort := m.awayClub.ShortName

	scoreLine := scoreStyle.Render(
		fmt.Sprintf("%s  %d – %d  %s", homeShort, s.HomeGoals, s.AwayGoals, awayShort),
	)

	statsLines := []string{
		titleStyle.Render("Match Result"),
		"",
		scoreLine,
		"",
		titleStyle.Render("Stats"),
		renderStatRow("Shots", s.Stats.Shots),
		renderStatRow("Goals", s.Stats.Goals),
		renderStatRow("Saves", s.Stats.Saves),
		renderStatRow("Blocks", s.Stats.Blocks),
		renderStatRow("Set Pieces", s.Stats.SetPieces),
		renderStatRow("Cards", s.Stats.Cards),
		renderStatRow("Injuries", s.Stats.Injuries),
		renderStatRow("Build-Ups", s.Stats.BuildUps),
		"",
		titleStyle.Render("Key Events"),
	}

	keyEvents := filterKeyEvents(s.Events)
	total := len(keyEvents)
	end := m.eventScroll + m.maxEventLines
	if end > total {
		end = total
	}
	visible := keyEvents[m.eventScroll:end]

	for _, event := range visible {
		statsLines = append(statsLines, renderEventLine(event))
	}

	if total > m.maxEventLines {
		scrollInfo := fmt.Sprintf("  %d–%d / %d events  (j/k to scroll)",
			m.eventScroll+1, end, total)
		statsLines = append(statsLines, "", dimStyle.Render(scrollInfo))
	} else if total == 0 {
		statsLines = append(statsLines, dimStyle.Render("  No key events recorded."))
	}

	return strings.Join(statsLines, "\n")
}

func renderStatRow(label string, value int) string {
	return lipgloss.JoinHorizontal(lipgloss.Top,
		statLabelStyle.Render("  "+label),
		statValueStyle.Render(fmt.Sprintf("%d", value)),
	)
}

func filterKeyEvents(events []domain.MatchEvent) []domain.MatchEvent {
	key := []domain.MatchEvent{}
	for _, e := range events {
		switch e.Type {
		case "goal", "yellow", "red", "injury", "penalty", "corner", "free_kick":
			key = append(key, e)
		}
	}
	return key
}

func renderEventLine(event domain.MatchEvent) string {
	minute := fmt.Sprintf("%3d'", event.Minute)
	label := eventLabel(event.Type)
	line := fmt.Sprintf("  %s  %s", minute, label)
	switch event.Type {
	case "goal":
		return eventGoalStyle.Render(line)
	case "yellow", "red":
		return eventCardStyle.Render(line)
	case "injury":
		return eventInjuryStyle.Render(line)
	default:
		return eventDefaultStyle.Render(line)
	}
}

func eventLabel(eventType string) string {
	switch eventType {
	case "goal":
		return "GOAL"
	case "yellow":
		return "Yellow card"
	case "red":
		return "Red card"
	case "injury":
		return "Injury"
	case "penalty":
		return "Penalty awarded"
	case "corner":
		return "Corner"
	case "free_kick":
		return "Free kick"
	default:
		return eventType
	}
}

// ---------------------------------------------------------------------------
// Demo data helpers
// ---------------------------------------------------------------------------

func demoTeam(id int64, name, shortName string) (domain.Club, []domain.Player) {
	club := domain.Club{
		ID:        id,
		BranchID:  1,
		ExtKey:    shortName,
		Name:      name,
		ShortName: shortName,
	}
	squad := buildDemoSquad(id*100, shortName)
	return club, squad
}

// buildDemoSquad returns a 20-player squad that covers all positions.
// IDs are offset by base to keep the two teams distinct.
func buildDemoSquad(base int64, team string) []domain.Player {
	type spec struct {
		name string
		pos  domain.Position
		ovr  int
	}
	specs := []spec{
		// GKs
		{"Goalkeeper A", domain.PositionGK, 82},
		{"Goalkeeper B", domain.PositionGK, 72},
		// Defenders
		{"Right Back A", domain.PositionRB, 78},
		{"Right Back B", domain.PositionRB, 71},
		{"Centre Back A", domain.PositionCB, 84},
		{"Centre Back B", domain.PositionCB, 81},
		{"Centre Back C", domain.PositionCB, 76},
		{"Left Back A", domain.PositionLB, 77},
		{"Left Back B", domain.PositionLB, 70},
		// Midfielders
		{"Defensive Mid A", domain.PositionDM, 80},
		{"Central Mid A", domain.PositionCM, 83},
		{"Central Mid B", domain.PositionCM, 79},
		{"Attacking Mid A", domain.PositionAM, 85},
		{"Attacking Mid B", domain.PositionAM, 73},
		// Wingers
		{"Right Wing A", domain.PositionRW, 82},
		{"Left Wing A", domain.PositionLW, 81},
		{"Left Wing B", domain.PositionLW, 74},
		// Strikers
		{"Striker A", domain.PositionST, 88},
		{"Striker B", domain.PositionST, 80},
		{"Striker C", domain.PositionST, 72},
	}

	players := make([]domain.Player, len(specs))
	for i, s := range specs {
		cid := base / 100
		players[i] = domain.Player{
			ID:              base + int64(i+1),
			ClubID:          &cid,
			FirstName:       team,
			LastName:        s.name,
			PrimaryPosition: s.pos,
			Attributes: domain.PlayerAttributes{
				Overall:   s.ovr,
				Pace:      s.ovr - 4 + i%5,
				Shooting:  clampAttr(s.ovr - 3 + i%7),
				Passing:   clampAttr(s.ovr - 2 + i%6),
				Defending: clampAttr(s.ovr - 5 + i%8),
				Keeping:   clampAttr(defensiveStat(s.pos, s.ovr, i)),
			},
		}
	}
	return players
}

func clampAttr(v int) int {
	if v < 40 {
		return 40
	}
	if v > 99 {
		return 99
	}
	return v
}

func defensiveStat(pos domain.Position, ovr, idx int) int {
	if pos == domain.PositionGK {
		return ovr + 5
	}
	return 40 + idx%15
}
