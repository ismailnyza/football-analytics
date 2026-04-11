// Package matchlab provides the Match Lab interactive screen for the TUI.
// It lets the user configure two demo teams, pick a formation, set a seed,
// run a deterministic simulation, and browse the resulting score, stats, and
// event log — all without leaving the terminal.
package matchlab

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ismael/football-analytics/internal/app"
	"github.com/ismael/football-analytics/internal/domain"
	"github.com/ismael/football-analytics/internal/engine/match"
	"github.com/ismael/football-analytics/internal/services/simulation"
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
// Pane and view-mode identifiers
// ---------------------------------------------------------------------------

type pane int

const (
	paneSetup  pane = 0
	paneResult pane = 1
)

// viewModeType distinguishes the three display layouts within Match Lab.
type viewModeType int

const (
	viewModeSplit  viewModeType = 0 // default: setup + result split panes
	viewModeResult viewModeType = 1 // full-screen match result detail
	viewModeEvents viewModeType = 2 // full-screen scrollable event log
)

// eventFilterType controls which events are shown in the event log view.
type eventFilterType int

const (
	filterAll       eventFilterType = 0
	filterGoals     eventFilterType = 1
	filterCards     eventFilterType = 2
	filterInjuries  eventFilterType = 3
	filterSetPieces eventFilterType = 4
)

var filterLabels = []string{"All", "Goals", "Cards", "Injuries", "Set Pieces"}

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

// Model is the Bubble Tea model for the Match Lab screen.
type Model struct {
	cfg             app.Config
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
	eligibleClubs   []domain.Club
	clubSquads      map[int64][]domain.Player
	homeIndex       int
	awayIndex       int
	history         []persistedMatch
	historyCursor   int
	eventScroll     int
	maxEventLines   int
	width           int
	height          int
	// sub-view state
	viewMode     viewModeType
	homePlan     match.TeamPlan
	awayPlan     match.TeamPlan
	eventFilter  eventFilterType
	resultScroll int
	status       string
}

type persistedMatch struct {
	Season  domain.Season
	Fixture domain.Fixture
	Match   domain.Match
	Home    domain.Club
	Away    domain.Club
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

// NewWithConfig returns a match lab model backed by the active branch when possible.
func NewWithConfig(cfg app.Config) Model {
	m := New()
	m.cfg = cfg

	state, err := app.LoadActiveBranchState(context.Background(), cfg)
	if err != nil {
		m.status = "using demo teams: " + err.Error()
		return m
	}
	eligible := state.EligibleClubs(11)
	if len(eligible) < 2 {
		m.status = "using demo teams: need 2 clubs with 11 players"
		return m
	}

	homeClub := eligible[0]
	awayClub := eligible[1]
	for _, club := range eligible {
		if club.ID == cfg.CurrentClubID() {
			homeClub = club
			break
		}
	}
	for _, club := range eligible {
		if club.ID != homeClub.ID {
			awayClub = club
			break
		}
	}

	m.homeClub = homeClub
	m.awayClub = awayClub
	m.eligibleClubs = append([]domain.Club(nil), eligible...)
	m.clubSquads = state.Squads
	m.homeSquad = append([]domain.Player(nil), state.Squads[homeClub.ID]...)
	m.awaySquad = append([]domain.Player(nil), state.Squads[awayClub.ID]...)
	m.homeIndex = indexOfClub(eligible, homeClub.ID)
	m.awayIndex = indexOfClub(eligible, awayClub.ID)
	m.history = loadPersistedMatches(context.Background(), cfg)
	m.status = fmt.Sprintf("storage-backed squads: %s vs %s", homeClub.ShortName, awayClub.ShortName)
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	kMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch m.viewMode {
	case viewModeResult:
		switch kMsg.String() {
		case "b":
			m.viewMode = viewModeSplit
		case "e":
			m.viewMode = viewModeEvents
			m.eventScroll = 0
		case "j", "down":
			m.resultScroll++
		case "k", "up":
			if m.resultScroll > 0 {
				m.resultScroll--
			}
		}

	case viewModeEvents:
		switch kMsg.String() {
		case "b":
			m.viewMode = viewModeSplit
		case "v":
			m.viewMode = viewModeResult
			m.resultScroll = 0
		case "j", "down":
			if m.simulated {
				filtered := m.filteredEvents()
				maxScroll := len(filtered) - m.maxEventLines
				if maxScroll < 0 {
					maxScroll = 0
				}
				if m.eventScroll < maxScroll {
					m.eventScroll++
				}
			}
		case "k", "up":
			if m.eventScroll > 0 {
				m.eventScroll--
			}
		case "f":
			m.eventFilter = (m.eventFilter + 1) % eventFilterType(len(filterLabels))
			m.eventScroll = 0
		}

	default: // viewModeSplit
		switch kMsg.String() {
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
		case "[":
			if m.activePaneFocus == paneSetup {
				m = m.cycleHome(-1)
			}
		case "]":
			if m.activePaneFocus == paneSetup {
				m = m.cycleHome(1)
			}
		case "{":
			if m.activePaneFocus == paneSetup {
				m = m.cycleAway(-1)
			}
		case "}":
			if m.activePaneFocus == paneSetup {
				m = m.cycleAway(1)
			}
		case "up", "k":
			if m.activePaneFocus == paneSetup {
				if m.historyCursor > 0 {
					m.historyCursor--
				}
			} else if m.eventScroll > 0 {
				m.eventScroll--
			}
		case "down", "j":
			if m.activePaneFocus == paneSetup {
				if m.historyCursor < len(m.history)-1 {
					m.historyCursor++
				}
			} else if m.simulated {
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
		case "p":
			if m.activePaneFocus == paneSetup {
				m = m.loadSelectedHistory()
			}
		case "c":
			if m.activePaneFocus == paneSetup {
				m = m.createBranchFromSelectedMatch()
			}
		case "r":
			m.simulated = false
			m.summary = nil
			m.eventScroll = 0
			m.resultScroll = 0
			m.status = "results cleared"
		case "+", "=":
			m.seed++
		case "-":
			if m.seed > 0 {
				m.seed--
			}
		case "v":
			if m.simulated {
				m.viewMode = viewModeResult
				m.resultScroll = 0
			}
		case "e":
			if m.simulated {
				m.viewMode = viewModeEvents
				m.eventScroll = 0
			}
		}
	}
	return m, nil
}

func (m Model) runSimulation() Model {
	formation := m.formations[m.formationIdx]
	homeLineup, err := match.SelectLineup(m.homeSquad, formation, nil)
	if err != nil {
		m.status = err.Error()
		return m
	}
	awayLineup, err := match.SelectLineup(m.awaySquad, formation, nil)
	if err != nil {
		m.status = err.Error()
		return m
	}
	homePlan := match.BuildTeamPlan(m.homeClub, homeLineup)
	awayPlan := match.BuildTeamPlan(m.awayClub, awayLineup)
	summary := match.RunTickLoop(m.seed, homePlan, awayPlan, 0)
	m.summary = &summary
	m.simulated = true
	m.eventScroll = 0
	m.homePlan = homePlan
	m.awayPlan = awayPlan
	m.status = fmt.Sprintf("simulated %s vs %s", m.homeClub.ShortName, m.awayClub.ShortName)
	if m.cfg.Repo != nil && m.cfg.CurrentBranchID() != 0 {
		if err := m.persistSimulation(summary); err != nil {
			m.status = "persist failed: " + err.Error()
		}
		m.history = loadPersistedMatches(context.Background(), m.cfg)
	}
	return m
}

func (m Model) cycleHome(delta int) Model {
	if len(m.eligibleClubs) < 2 {
		return m
	}
	next := normalizeClubIndex(m.homeIndex+delta, len(m.eligibleClubs))
	if next == m.awayIndex {
		next = normalizeClubIndex(next+delta, len(m.eligibleClubs))
	}
	m.homeIndex = next
	m.homeClub = m.eligibleClubs[m.homeIndex]
	m.homeSquad = append([]domain.Player(nil), m.clubSquads[m.homeClub.ID]...)
	m.status = fmt.Sprintf("home team: %s", m.homeClub.Name)
	return m
}

func (m Model) cycleAway(delta int) Model {
	if len(m.eligibleClubs) < 2 {
		return m
	}
	next := normalizeClubIndex(m.awayIndex+delta, len(m.eligibleClubs))
	if next == m.homeIndex {
		next = normalizeClubIndex(next+delta, len(m.eligibleClubs))
	}
	m.awayIndex = next
	m.awayClub = m.eligibleClubs[m.awayIndex]
	m.awaySquad = append([]domain.Player(nil), m.clubSquads[m.awayClub.ID]...)
	m.status = fmt.Sprintf("away team: %s", m.awayClub.Name)
	return m
}

func (m Model) loadSelectedHistory() Model {
	if len(m.history) == 0 || m.historyCursor >= len(m.history) {
		return m
	}
	entry := m.history[m.historyCursor]
	events, err := m.cfg.Repo.ListMatchEvents(context.Background(), entry.Match.ID)
	if err != nil {
		m.status = "history load failed: " + err.Error()
		return m
	}
	summary := summaryFromStoredMatch(entry.Match, events, entry.Home, entry.Away)
	m.homeClub = entry.Home
	m.awayClub = entry.Away
	if squad, ok := m.clubSquads[entry.Home.ID]; ok {
		m.homeSquad = append([]domain.Player(nil), squad...)
	}
	if squad, ok := m.clubSquads[entry.Away.ID]; ok {
		m.awaySquad = append([]domain.Player(nil), squad...)
	}
	m.summary = &summary
	m.simulated = true
	m.eventScroll = 0
	m.resultScroll = 0
	m.status = fmt.Sprintf("loaded saved match %d from %q", entry.Match.ID, entry.Season.Label)
	return m
}

func (m Model) createBranchFromSelectedMatch() Model {
	if len(m.history) == 0 || m.historyCursor >= len(m.history) || m.cfg.Repo == nil || m.cfg.CurrentWorldID() == 0 {
		return m
	}
	entry := m.history[m.historyCursor]
	branchName := fmt.Sprintf("match-%d-replay", entry.Match.ID)
	branch, err := m.cfg.Repo.CreateBranch(context.Background(), domain.Branch{
		WorldID:            m.cfg.CurrentWorldID(),
		ParentBranchID:     ptrInt64(m.cfg.CurrentBranchID()),
		Name:               branchName,
		CreatedFromMatchID: ptrInt64(entry.Match.ID),
	})
	if err != nil {
		m.status = "branch create failed: " + err.Error()
		return m
	}
	if m.cfg.State != nil {
		m.cfg.State.ActiveBranchID = branch.ID
		m.cfg.State.ActiveBranchName = branch.Name
	}
	m.status = fmt.Sprintf("created branch %q from match %d", branch.Name, entry.Match.ID)
	return m
}

func (m *Model) persistSimulation(summary match.Summary) error {
	ctx := context.Background()
	seasonRecord, err := m.cfg.Repo.CreateSeason(ctx, domain.Season{
		BranchID:  m.cfg.CurrentBranchID(),
		Label:     fmt.Sprintf("Match Lab %d %d", m.seed, time.Now().UTC().UnixNano()),
		StartYear: time.Now().UTC().Year(),
	})
	if err != nil {
		return err
	}
	fixtureRecord, err := m.cfg.Repo.CreateFixture(ctx, domain.Fixture{
		SeasonID:   seasonRecord.ID,
		Matchday:   1,
		HomeClubID: m.homeClub.ID,
		AwayClubID: m.awayClub.ID,
		Scheduled:  time.Now().UTC(),
	})
	if err != nil {
		return err
	}
	matchRecord, err := simulation.PersistMatchSummary(
		ctx,
		m.cfg.Repo,
		fixtureRecord.ID,
		m.cfg.CurrentBranchID(),
		m.seed,
		summary,
	)
	if err != nil {
		return err
	}
	m.status = fmt.Sprintf("saved match %d in season %q", matchRecord.ID, seasonRecord.Label)
	return nil
}

// ---------------------------------------------------------------------------
// View dispatcher
// ---------------------------------------------------------------------------

// View renders the current Match Lab layout based on the active view mode.
func (m Model) View(width, height int) string {
	m.width = width
	m.height = height
	switch m.viewMode {
	case viewModeResult:
		return m.renderFullResult(width, height)
	case viewModeEvents:
		return m.renderFullEventLog(width, height)
	default:
		return m.renderSplitView(width, height)
	}
}

// ---------------------------------------------------------------------------
// Split view (default)
// ---------------------------------------------------------------------------

func (m Model) renderSplitView(width, height int) string {
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

	hint := "tab: switch pane  s: simulate  p: replay  c: branch  r: reset  ←/→: formation  [ ] / { }: teams  +/-: seed  j/k: scroll"
	if m.simulated {
		hint += "  v: result detail  e: event log"
	}
	helpBar := helpBarStyle.Render(hint)

	return lipgloss.JoinVertical(lipgloss.Left, body, "", helpBar)
}

// ---------------------------------------------------------------------------
// Split pane renderers
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
	if len(m.history) > 0 {
		lines = append(lines, "",
			labelStyle.Render("Saved Matches"),
		)
		for i, entry := range m.history {
			row := fmt.Sprintf("  %-6s %d-%d %-6s  %s",
				entry.Home.ShortName, entry.Match.HomeGoals, entry.Match.AwayGoals, entry.Away.ShortName, truncate(entry.Season.Label, 18))
			if i == m.historyCursor {
				lines = append(lines, highlightStyle.Render(row))
			} else {
				lines = append(lines, dimStyle.Render(row))
			}
		}
		lines = append(lines, dimStyle.Render("  j/k select saved  p load replay"))
	}
	if m.status != "" {
		lines = append(lines, "", dimStyle.Render("Status: "+m.status))
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

// ---------------------------------------------------------------------------
// Full-screen result detail view (SEC-012)
// ---------------------------------------------------------------------------

func (m Model) renderFullResult(width, height int) string {
	if !m.simulated || m.summary == nil {
		return dimStyle.Render("No match result available. Press b to return.")
	}
	s := m.summary
	homeShort := m.homeClub.ShortName
	awayShort := m.awayClub.ShortName

	lines := []string{
		titleStyle.Render("Match Result — Detail"),
		"",
		scoreStyle.Render(fmt.Sprintf("%s  %d – %d  %s", homeShort, s.HomeGoals, s.AwayGoals, awayShort)),
		"",
		titleStyle.Render("Statistics"),
		renderStatRow("Build-Ups", s.Stats.BuildUps),
		renderStatRow("Penetrations", s.Stats.Penetrations),
		renderStatRow("Shots", s.Stats.Shots),
		renderStatRow("Goals", s.Stats.Goals),
		renderStatRow("Saves", s.Stats.Saves),
		renderStatRow("Blocks", s.Stats.Blocks),
		renderStatRow("Turnovers", s.Stats.Turnovers),
		renderStatRow("Set Pieces", s.Stats.SetPieces),
		renderStatRow("Cards", s.Stats.Cards),
		renderStatRow("Injuries", s.Stats.Injuries),
		"",
		titleStyle.Render("Team Condition"),
		labelStyle.Render(fmt.Sprintf("  %s  avg fatigue: %.2f", homeShort, s.HomeAverageFatigue)),
		labelStyle.Render(fmt.Sprintf("  %s  avg fatigue: %.2f", awayShort, s.AwayAverageFatigue)),
	}

	if len(s.Injuries) > 0 {
		lines = append(lines, "", titleStyle.Render("Injuries"))
		for _, inj := range s.Injuries {
			lines = append(lines, eventInjuryStyle.Render(fmt.Sprintf(
				"  %3d'  %s  Player #%d  (%s)", cardMinute(inj.Tick), inj.Team, inj.PlayerID, inj.Severity,
			)))
		}
	}

	if len(s.Cards) > 0 {
		lines = append(lines, "", titleStyle.Render("Cards"))
		for _, card := range s.Cards {
			line := fmt.Sprintf("  %3d'  %s  Player #%d  (%s)", cardMinute(card.Tick), card.Team, card.PlayerID, card.Card)
			if card.Card == "red" {
				lines = append(lines, eventCardStyle.Render(line))
			} else {
				lines = append(lines, eventDefaultStyle.Render(line))
			}
		}
	}

	if len(s.Suspensions) > 0 {
		lines = append(lines, "", titleStyle.Render("Suspensions"))
		for _, sus := range s.Suspensions {
			lines = append(lines, valueStyle.Render(fmt.Sprintf(
				"  %s  Player #%d  — %d match(es)  (%s)", sus.Team, sus.PlayerID, sus.Matches, sus.Reason,
			)))
		}
	}

	lines = append(lines, "",
		dimStyle.Render(fmt.Sprintf(
			"  Seed: %d  |  Ticks: %d  |  Formation: %s",
			m.seed, s.TotalTicks, m.formations[m.formationIdx],
		)),
	)

	// Scroll window
	bodyH := height - 4
	if bodyH < 4 {
		bodyH = 4
	}
	total := len(lines)
	maxScroll := total - bodyH
	if maxScroll < 0 {
		maxScroll = 0
	}
	start := m.resultScroll
	if start > maxScroll {
		start = maxScroll
	}
	end := start + bodyH
	if end > total {
		end = total
	}

	body := strings.Join(lines[start:end], "\n")
	if total > bodyH {
		body += "\n" + dimStyle.Render(fmt.Sprintf("  line %d–%d of %d  (j/k to scroll)", start+1, end, total))
	}

	helpBar := helpBarStyle.Render("b: back to setup  e: event log  j/k: scroll")
	return lipgloss.JoinVertical(lipgloss.Left, body, "", helpBar)
}

// ---------------------------------------------------------------------------
// Full-screen event log view (SEC-012)
// ---------------------------------------------------------------------------

func (m Model) renderFullEventLog(width, height int) string {
	if !m.simulated || m.summary == nil {
		return dimStyle.Render("No match result available. Press b to return.")
	}
	s := m.summary
	homeShort := m.homeClub.ShortName
	awayShort := m.awayClub.ShortName

	// Filter bar
	filterParts := make([]string, len(filterLabels))
	for i, label := range filterLabels {
		if eventFilterType(i) == m.eventFilter {
			filterParts[i] = highlightStyle.Render(label)
		} else {
			filterParts[i] = dimStyle.Render(label)
		}
	}

	filtered := m.filteredEvents()
	total := len(filtered)

	bodyH := height - 9
	if bodyH < 4 {
		bodyH = 4
	}
	maxScroll := total - bodyH
	if maxScroll < 0 {
		maxScroll = 0
	}
	start := m.eventScroll
	if start > maxScroll {
		start = maxScroll
	}
	end := start + bodyH
	if end > total {
		end = total
	}

	eventLines := make([]string, 0, end-start)
	for _, event := range filtered[start:end] {
		eventLines = append(eventLines, renderEventLine(event))
	}

	parts := []string{
		titleStyle.Render("Event Log"),
		"",
		scoreStyle.Render(fmt.Sprintf("%s  %d – %d  %s", homeShort, s.HomeGoals, s.AwayGoals, awayShort)),
		"",
		"  " + strings.Join(filterParts, "  "),
		"",
	}
	parts = append(parts, eventLines...)

	if total > bodyH {
		parts = append(parts, "", dimStyle.Render(fmt.Sprintf(
			"  %d–%d / %d events  (j/k to scroll)", start+1, end, total,
		)))
	} else if total == 0 {
		parts = append(parts, dimStyle.Render("  No events match filter."))
	}

	helpBar := helpBarStyle.Render("b: back to setup  v: result detail  f: cycle filter  j/k: scroll")
	parts = append(parts, "", helpBar)

	return strings.Join(parts, "\n")
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

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

func (m Model) filteredEvents() []domain.MatchEvent {
	if m.summary == nil {
		return nil
	}
	if m.eventFilter == filterAll {
		return m.summary.Events
	}
	result := make([]domain.MatchEvent, 0, len(m.summary.Events)/4)
	for _, event := range m.summary.Events {
		if matchesEventFilter(m.eventFilter, event) {
			result = append(result, event)
		}
	}
	return result
}

func matchesEventFilter(f eventFilterType, event domain.MatchEvent) bool {
	switch f {
	case filterGoals:
		return event.Type == "goal"
	case filterCards:
		return event.Type == "yellow" || event.Type == "red"
	case filterInjuries:
		return event.Type == "injury"
	case filterSetPieces:
		return event.Type == "corner" || event.Type == "free_kick" || event.Type == "penalty"
	default:
		return true
	}
}

// cardMinute converts a tick number to a match minute.
func cardMinute(tick int) int {
	if tick <= 0 {
		return 0
	}
	return ((tick - 1) * 6) / 60
}

func loadPersistedMatches(ctx context.Context, cfg app.Config) []persistedMatch {
	if cfg.Repo == nil || cfg.CurrentBranchID() == 0 {
		return nil
	}
	state, err := app.LoadActiveBranchState(ctx, cfg)
	if err != nil {
		return nil
	}
	clubsByID := make(map[int64]domain.Club, len(state.Clubs))
	for _, club := range state.Clubs {
		clubsByID[club.ID] = club
	}
	seasons, err := cfg.Repo.ListSeasonsByBranch(ctx, cfg.CurrentBranchID())
	if err != nil {
		return nil
	}
	out := make([]persistedMatch, 0)
	for _, seasonRecord := range seasons {
		if !strings.HasPrefix(seasonRecord.Label, "Match Lab") {
			continue
		}
		fixtures, err := cfg.Repo.ListFixturesBySeason(ctx, seasonRecord.ID)
		if err != nil {
			continue
		}
		fixtureByID := make(map[int64]domain.Fixture, len(fixtures))
		for _, fixture := range fixtures {
			fixtureByID[fixture.ID] = fixture
		}
		matches, err := cfg.Repo.ListMatchesBySeason(ctx, seasonRecord.ID)
		if err != nil {
			continue
		}
		for _, matchRecord := range matches {
			fixture, ok := fixtureByID[matchRecord.FixtureID]
			if !ok {
				continue
			}
			out = append(out, persistedMatch{
				Season:  seasonRecord,
				Fixture: fixture,
				Match:   matchRecord,
				Home:    clubsByID[fixture.HomeClubID],
				Away:    clubsByID[fixture.AwayClubID],
			})
		}
	}
	return out
}

func summaryFromStoredMatch(matchRecord domain.Match, events []domain.MatchEvent, home domain.Club, away domain.Club) match.Summary {
	stats := match.MatchStats{}
	injuries := make([]match.Injury, 0)
	cards := make([]match.CardRecord, 0)
	suspensions := make([]match.Suspension, 0)
	for _, event := range events {
		switch event.Type {
		case "build_up":
			stats.BuildUps++
		case "penetration":
			stats.Penetrations++
		case "shot":
			stats.Shots++
		case "goal":
			stats.Goals++
		case "save":
			stats.Saves++
		case "block":
			stats.Blocks++
		case "turnover":
			stats.Turnovers++
		case "corner", "free_kick", "penalty":
			stats.SetPieces++
		case "yellow", "red":
			stats.Cards++
			cards = append(cards, match.CardRecord{
				PlayerID: derefInt64(event.PlayerID),
				Team:     eventTeam(event, home, away),
				Tick:     event.Tick,
				Card:     event.Type,
			})
			if event.Type == "red" {
				suspensions = append(suspensions, match.Suspension{
					PlayerID: derefInt64(event.PlayerID),
					Team:     eventTeam(event, home, away),
					Reason:   "red_card",
					Matches:  1,
				})
			}
		case "injury":
			stats.Injuries++
			injuries = append(injuries, match.Injury{
				PlayerID: derefInt64(event.PlayerID),
				Team:     eventTeam(event, home, away),
				Tick:     event.Tick,
				Severity: "stored",
			})
		}
	}
	return match.Summary{
		TotalTicks:  matchRecord.TickCount,
		HomeGoals:   matchRecord.HomeGoals,
		AwayGoals:   matchRecord.AwayGoals,
		Stats:       stats,
		Injuries:    injuries,
		Cards:       cards,
		Suspensions: suspensions,
		Events:      events,
	}
}

func eventTeam(event domain.MatchEvent, home domain.Club, away domain.Club) string {
	if event.TeamClubID == nil {
		return ""
	}
	switch *event.TeamClubID {
	case home.ID:
		return home.ShortName
	case away.ID:
		return away.ShortName
	default:
		return fmt.Sprintf("%d", *event.TeamClubID)
	}
}

func derefInt64(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func indexOfClub(clubs []domain.Club, clubID int64) int {
	for i, club := range clubs {
		if club.ID == clubID {
			return i
		}
	}
	return 0
}

func normalizeClubIndex(index, size int) int {
	if size == 0 {
		return 0
	}
	for index < 0 {
		index += size
	}
	return index % size
}

func ptrInt64(v int64) *int64 {
	if v == 0 {
		return nil
	}
	value := v
	return &value
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
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
