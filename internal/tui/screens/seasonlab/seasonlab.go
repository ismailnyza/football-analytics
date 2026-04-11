// Package seasonlab provides the Season Lab interactive TUI screen.
// It lets the user configure a league of demo clubs, run a full season
// simulation, and browse the resulting standings and match results.
package seasonlab

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
	"github.com/ismael/football-analytics/internal/engine/season"
	"github.com/ismael/football-analytics/internal/services/simulation"
)

// ---------------------------------------------------------------------------
// Styles (re-use the same palette as Match Lab for consistency)
// ---------------------------------------------------------------------------

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("117"))
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	hlStyle    = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("62")).
			Padding(0, 1)
	scoreStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("220"))
	helpBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	paneStyle    = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2)
	activePaneStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)
)

// ---------------------------------------------------------------------------
// View modes
// ---------------------------------------------------------------------------

type viewMode int

const (
	viewSetup     viewMode = 0
	viewStandings viewMode = 1
	viewResults   viewMode = 2
)

// ---------------------------------------------------------------------------
// Model
// ---------------------------------------------------------------------------

// Model is the Bubble Tea model for the Season Lab screen.
type Model struct {
	cfg           app.Config
	mode          viewMode
	seed          int64
	simulated     bool
	result        *season.SeasonResult
	tableScroll   int
	resScroll     int
	width         int
	height        int
	status        string
	clubCount     int
	history       []persistedSeason
	historyCursor int
}

type persistedSeason struct {
	Season     domain.Season
	MatchCount int
}

// New returns an initialised Season Lab model.
func New() Model {
	return Model{
		mode:      viewSetup,
		seed:      1,
		clubCount: 6,
	}
}

// NewWithConfig returns a season lab model backed by the active branch when possible.
func NewWithConfig(cfg app.Config) Model {
	m := New()
	m.cfg = cfg
	state, err := app.LoadActiveBranchState(context.Background(), cfg)
	if err != nil {
		m.status = "using demo league: " + err.Error()
		return m
	}
	eligible := state.EligibleClubs(11)
	if len(eligible) < 2 {
		m.status = "using demo league: need 2 clubs with 11 players"
		return m
	}
	m.clubCount = len(eligible)
	m.status = fmt.Sprintf("storage-backed league loaded: %d clubs", len(eligible))
	m.history = loadPersistedSeasons(context.Background(), cfg)
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	kMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch m.mode {
	case viewStandings:
		switch kMsg.String() {
		case "b":
			m.mode = viewSetup
		case "r":
			m.mode = viewResults
			m.resScroll = 0
		case "j", "down":
			m.tableScroll++
		case "k", "up":
			if m.tableScroll > 0 {
				m.tableScroll--
			}
		}
	case viewResults:
		switch kMsg.String() {
		case "b":
			m.mode = viewSetup
		case "t":
			m.mode = viewStandings
		case "j", "down":
			m.resScroll++
		case "k", "up":
			if m.resScroll > 0 {
				m.resScroll--
			}
		}
	default: // viewSetup
		switch kMsg.String() {
		case "s":
			m = m.runSeason()
		case "j", "down":
			if m.historyCursor < len(m.history)-1 {
				m.historyCursor++
			}
		case "k", "up":
			if m.historyCursor > 0 {
				m.historyCursor--
			}
		case "h":
			m = m.loadSelectedHistory()
		case "c":
			m = m.createBranchFromSelectedSeason()
		case "r":
			m.simulated = false
			m.result = nil
			m.tableScroll = 0
			m.resScroll = 0
			m.status = "season results cleared"
		case "+", "=":
			m.seed++
		case "-":
			if m.seed > 0 {
				m.seed--
			}
		case "t":
			if m.simulated {
				m.mode = viewStandings
			}
		case "v":
			if m.simulated {
				m.mode = viewResults
			}
		}
	}
	return m, nil
}

func (m Model) loadSelectedHistory() Model {
	if len(m.history) == 0 || m.historyCursor >= len(m.history) {
		return m
	}
	result, err := loadPersistedSeasonResult(context.Background(), m.cfg, m.history[m.historyCursor].Season)
	if err != nil {
		m.status = "history load failed: " + err.Error()
		return m
	}
	m.result = &result
	m.simulated = true
	m.tableScroll = 0
	m.resScroll = 0
	m.status = fmt.Sprintf("loaded saved season %q", result.Season.Label)
	return m
}

func (m Model) createBranchFromSelectedSeason() Model {
	if len(m.history) == 0 || m.historyCursor >= len(m.history) || m.cfg.Repo == nil || m.cfg.CurrentWorldID() == 0 {
		return m
	}
	entry := m.history[m.historyCursor]
	branchName := fmt.Sprintf("season-%d-replay", entry.Season.ID)
	branch, err := m.cfg.Repo.CreateBranch(context.Background(), domain.Branch{
		WorldID:        m.cfg.CurrentWorldID(),
		ParentBranchID: ptrInt64(m.cfg.CurrentBranchID()),
		Name:           branchName,
	})
	if err != nil {
		m.status = "branch create failed: " + err.Error()
		return m
	}
	if m.cfg.State != nil {
		m.cfg.State.ActiveBranchID = branch.ID
		m.cfg.State.ActiveBranchName = branch.Name
	}
	m.status = fmt.Sprintf("created branch %q from season %q", branch.Name, entry.Season.Label)
	return m
}

func (m Model) runSeason() Model {
	clubs, squads := demoLeague()
	if m.cfg.Repo != nil && m.cfg.CurrentBranchID() != 0 {
		if state, err := app.LoadActiveBranchState(context.Background(), m.cfg); err == nil {
			eligible := state.EligibleClubs(11)
			if len(eligible) >= 2 {
				clubs = eligible
				squads = make(map[int64][]domain.Player, len(eligible))
				for _, club := range eligible {
					squads[club.ID] = append([]domain.Player(nil), state.Squads[club.ID]...)
				}
				m.clubCount = len(eligible)
			} else {
				m.status = "using demo league: need 2 clubs with 11 players"
			}
		} else {
			m.status = "using demo league: " + err.Error()
		}
	}
	clubMap := make(map[int64]domain.Club, len(clubs))
	squadMap := make(map[int64][]domain.Player, len(clubs))
	for _, c := range clubs {
		clubMap[c.ID] = c
		squadMap[c.ID] = squads[c.ID]
	}

	seasonRecord := domain.Season{ID: 1, BranchID: 1, Label: "Demo Season", StartYear: 2024}
	if m.cfg.CurrentBranchID() != 0 {
		seasonRecord.BranchID = m.cfg.CurrentBranchID()
		seasonRecord.StartYear = time.Now().UTC().Year()
		seasonRecord.Label = fmt.Sprintf("Season Lab %d %d", m.seed, time.Now().UTC().UnixNano())
	}
	fixtures, err := season_generateFixtures(seasonRecord, clubs)
	if err != nil {
		m.status = err.Error()
		return m
	}

	if m.cfg.Repo != nil && m.cfg.CurrentBranchID() != 0 {
		persistedSeason, persistedFixtures, err := m.persistSeasonShell(seasonRecord, fixtures)
		if err != nil {
			m.status = "persist failed: " + err.Error()
			return m
		}
		seasonRecord = persistedSeason
		fixtures = persistedFixtures
	}

	result := season_run(m.seed, seasonRecord, fixtures, squadMap, clubMap)
	if m.cfg.Repo != nil && m.cfg.CurrentBranchID() != 0 {
		if err := m.persistSeasonResults(result); err != nil {
			m.status = "match persist failed: " + err.Error()
			return m
		}
		m.status = fmt.Sprintf("saved season %q with %d matches", result.Season.Label, len(result.Results))
		m.history = loadPersistedSeasons(context.Background(), m.cfg)
	} else {
		m.status = fmt.Sprintf("simulated %d matches", len(result.Results))
	}
	m.result = &result
	m.simulated = true
	m.tableScroll = 0
	m.resScroll = 0
	return m
}

func (m Model) persistSeasonShell(seasonRecord domain.Season, fixtures []domain.Fixture) (domain.Season, []domain.Fixture, error) {
	ctx := context.Background()
	persistedSeason, err := m.cfg.Repo.CreateSeason(ctx, seasonRecord)
	if err != nil {
		return domain.Season{}, nil, err
	}
	persistedFixtures := make([]domain.Fixture, 0, len(fixtures))
	for _, fixture := range fixtures {
		fixture.SeasonID = persistedSeason.ID
		persistedFixture, err := m.cfg.Repo.CreateFixture(ctx, fixture)
		if err != nil {
			return domain.Season{}, nil, err
		}
		persistedFixtures = append(persistedFixtures, persistedFixture)
	}
	return persistedSeason, persistedFixtures, nil
}

func (m Model) persistSeasonResults(result season.SeasonResult) error {
	ctx := context.Background()
	for idx, matchResult := range result.Results {
		matchSeed := m.seed ^ int64(idx+1)
		if _, err := simulation.PersistMatchSummary(
			ctx,
			m.cfg.Repo,
			matchResult.Fixture.ID,
			m.cfg.CurrentBranchID(),
			matchSeed,
			matchResult.Summary,
		); err != nil {
			return err
		}
	}
	return nil
}

func season_generateFixtures(s domain.Season, clubs []domain.Club) ([]domain.Fixture, error) {
	fixtures, err := season.GenerateFixtures(s, clubs, time.Time{})
	if err != nil {
		return nil, err
	}
	for i := range fixtures {
		fixtures[i].ID = int64(i + 1)
	}
	return fixtures, nil
}

func season_run(seed int64, s domain.Season, fixtures []domain.Fixture, squads map[int64][]domain.Player, clubs map[int64]domain.Club) season.SeasonResult {
	return season.RunSeason(seed, s, fixtures, squads, clubs, "4-3-3")
}

func loadPersistedSeasons(ctx context.Context, cfg app.Config) []persistedSeason {
	if cfg.Repo == nil || cfg.CurrentBranchID() == 0 {
		return nil
	}
	seasons, err := cfg.Repo.ListSeasonsByBranch(ctx, cfg.CurrentBranchID())
	if err != nil {
		return nil
	}
	out := make([]persistedSeason, 0, len(seasons))
	for _, seasonRecord := range seasons {
		if !strings.HasPrefix(seasonRecord.Label, "Season Lab") {
			continue
		}
		matches, err := cfg.Repo.ListMatchesBySeason(ctx, seasonRecord.ID)
		if err != nil {
			continue
		}
		out = append(out, persistedSeason{Season: seasonRecord, MatchCount: len(matches)})
	}
	return out
}

func loadPersistedSeasonResult(ctx context.Context, cfg app.Config, seasonRecord domain.Season) (season.SeasonResult, error) {
	state, err := app.LoadActiveBranchState(ctx, cfg)
	if err != nil {
		return season.SeasonResult{}, err
	}
	fixtures, err := cfg.Repo.ListFixturesBySeason(ctx, seasonRecord.ID)
	if err != nil {
		return season.SeasonResult{}, err
	}
	matches, err := cfg.Repo.ListMatchesBySeason(ctx, seasonRecord.ID)
	if err != nil {
		return season.SeasonResult{}, err
	}
	matchByFixture := make(map[int64]domain.Match, len(matches))
	for _, matchRecord := range matches {
		matchByFixture[matchRecord.FixtureID] = matchRecord
	}
	results := make([]season.MatchResult, 0, len(fixtures))
	for _, fixture := range fixtures {
		matchRecord, ok := matchByFixture[fixture.ID]
		if !ok {
			continue
		}
		results = append(results, season.MatchResult{
			Fixture: fixture,
			Summary: match.Summary{
				TotalTicks: matchRecord.TickCount,
				HomeGoals:  matchRecord.HomeGoals,
				AwayGoals:  matchRecord.AwayGoals,
				Stats:      match.MatchStats{Goals: matchRecord.HomeGoals + matchRecord.AwayGoals},
			},
		})
	}
	return season.SeasonResult{
		Season:   seasonRecord,
		Fixtures: fixtures,
		Results:  results,
		Table:    season.ComputeTable(state.Clubs, fixtures, matches),
	}, nil
}

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

func (m Model) View(width, height int) string {
	m.width = width
	m.height = height
	switch m.mode {
	case viewStandings:
		return m.renderStandings(width, height)
	case viewResults:
		return m.renderResults(width, height)
	default:
		return m.renderSetup(width, height)
	}
}

func (m Model) renderSetup(width, height int) string {
	lines := []string{
		titleStyle.Render("Season Lab"),
		"",
		labelStyle.Render("League"),
		valueStyle.Render(fmt.Sprintf("  Active League  (%d clubs, 4-3-3)", m.clubCount)),
		"",
		labelStyle.Render("Seed"),
		valueStyle.Render(fmt.Sprintf("  %d", m.seed)),
		dimStyle.Render("  (+/- to change seed)"),
		"",
		hlStyle.Render(" [s] Simulate Season "),
	}
	if m.simulated && m.result != nil {
		lines = append(lines, "",
			dimStyle.Render(fmt.Sprintf("  %d matches simulated", len(m.result.Results))),
			dimStyle.Render("  t: standings  v: results  r: reset"),
		)
	}
	if len(m.history) > 0 {
		lines = append(lines, "", titleStyle.Render("Saved Seasons"))
		for i, entry := range m.history {
			row := fmt.Sprintf("  %-22s  %3d matches", truncate(entry.Season.Label, 22), entry.MatchCount)
			if i == m.historyCursor {
				lines = append(lines, hlStyle.Render(row))
			} else {
				lines = append(lines, dimStyle.Render(row))
			}
		}
		lines = append(lines, dimStyle.Render("  j/k select  h load saved season"))
	}
	if m.status != "" {
		lines = append(lines, "", dimStyle.Render("  "+m.status))
	}
	helpBar := helpBarStyle.Render("s: simulate  h: load saved  c: branch  +/-: seed  t: standings  v: results  r: reset")
	return lipgloss.JoinVertical(lipgloss.Left,
		strings.Join(lines, "\n"), "", helpBar)
}

func (m Model) renderStandings(width, height int) string {
	if m.result == nil {
		return dimStyle.Render("No season simulated.")
	}
	header := []string{
		titleStyle.Render("Season Standings"),
		"",
		labelStyle.Render(fmt.Sprintf("  %-20s %3s %3s %3s %3s %3s %4s %4s %4s",
			"Club", "P", "W", "D", "L", "Pts", "GF", "GA", "GD")),
	}
	rows := make([]string, len(m.result.Table))
	for i, s := range m.result.Table {
		rows[i] = fmt.Sprintf("  %2d. %-18s %3d %3d %3d %3d %3d %4d %4d %4d",
			i+1, truncate(s.ClubName, 18), s.Played, s.Won, s.Drawn, s.Lost,
			s.Points, s.GoalsFor, s.GoalsAgainst, s.GoalDifference)
	}

	bodyH := height - 6
	if bodyH < 4 {
		bodyH = 4
	}
	allLines := append(header, rows...)
	total := len(allLines)
	start := clamp(m.tableScroll, 0, max0(total-bodyH))
	end := min(start+bodyH, total)

	body := strings.Join(allLines[start:end], "\n")
	helpBar := helpBarStyle.Render("b: back  r: results  j/k: scroll")
	return lipgloss.JoinVertical(lipgloss.Left, body, "", helpBar)
}

func (m Model) renderResults(width, height int) string {
	if m.result == nil {
		return dimStyle.Render("No season simulated.")
	}
	header := []string{
		titleStyle.Render("Match Results"),
		"",
	}
	rows := make([]string, len(m.result.Results))
	for i, r := range m.result.Results {
		homeID := r.Fixture.HomeClubID
		awayID := r.Fixture.AwayClubID
		homeShort := clubShortName(m.result, homeID)
		awayShort := clubShortName(m.result, awayID)
		rows[i] = scoreStyle.Render(fmt.Sprintf(
			"  MD%2d  %-6s  %d – %d  %-6s",
			r.Fixture.Matchday, homeShort, r.Summary.HomeGoals, r.Summary.AwayGoals, awayShort,
		))
	}

	bodyH := height - 6
	if bodyH < 4 {
		bodyH = 4
	}
	allLines := append(header, rows...)
	total := len(allLines)
	start := clamp(m.resScroll, 0, max0(total-bodyH))
	end := min(start+bodyH, total)

	body := strings.Join(allLines[start:end], "\n")
	helpBar := helpBarStyle.Render("b: back  t: standings  j/k: scroll")
	return lipgloss.JoinVertical(lipgloss.Left, body, "", helpBar)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func clubShortName(result *season.SeasonResult, clubID int64) string {
	if result == nil {
		return "???"
	}
	for _, s := range result.Table {
		if s.ClubID == clubID {
			if len(s.ClubName) > 3 {
				return s.ClubName[:3]
			}
			return s.ClubName
		}
	}
	return "???"
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
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

func max0(v int) int {
	if v < 0 {
		return 0
	}
	return v
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func ptrInt64(v int64) *int64 {
	if v == 0 {
		return nil
	}
	value := v
	return &value
}

// ---------------------------------------------------------------------------
// Demo league data
// ---------------------------------------------------------------------------

func demoLeague() ([]domain.Club, map[int64][]domain.Player) {
	clubDefs := []struct {
		id   int64
		name string
		abbr string
	}{
		{1, "Arsenal FC", "ARS"},
		{2, "Liverpool FC", "LIV"},
		{3, "Manchester City", "MCI"},
		{4, "Chelsea FC", "CHE"},
		{5, "Tottenham", "TOT"},
		{6, "Newcastle Utd", "NEW"},
	}
	clubs := make([]domain.Club, len(clubDefs))
	squads := make(map[int64][]domain.Player, len(clubDefs))
	for i, def := range clubDefs {
		clubs[i] = domain.Club{ID: def.id, BranchID: 1, Name: def.name, ShortName: def.abbr}
		squads[def.id] = buildSquad(def.id, def.abbr)
	}
	return clubs, squads
}

func buildSquad(clubID int64, _ string) []domain.Player {
	type spec struct {
		pos domain.Position
		ovr int
	}
	specs := []spec{
		{domain.PositionGK, 80}, {domain.PositionGK, 70},
		{domain.PositionRB, 75}, {domain.PositionCB, 82}, {domain.PositionCB, 80}, {domain.PositionLB, 76},
		{domain.PositionDM, 79}, {domain.PositionCM, 81}, {domain.PositionCM, 78},
		{domain.PositionRW, 80}, {domain.PositionLW, 79}, {domain.PositionST, 83},
		{domain.PositionAM, 77}, {domain.PositionCB, 74}, {domain.PositionRB, 71},
		{domain.PositionLB, 71}, {domain.PositionDM, 73}, {domain.PositionCM, 72},
		{domain.PositionST, 75}, {domain.PositionST, 71},
	}
	players := make([]domain.Player, len(specs))
	for i, s := range specs {
		cid := clubID
		players[i] = domain.Player{
			ID:              clubID*100 + int64(i+1),
			ClubID:          &cid,
			PrimaryPosition: s.pos,
			Attributes: domain.PlayerAttributes{
				Overall:   s.ovr,
				Pace:      s.ovr - 3,
				Shooting:  s.ovr - 5,
				Passing:   s.ovr - 4,
				Defending: s.ovr - 6,
				Keeping:   keepStat(s.pos, s.ovr),
			},
		}
	}
	return players
}

func keepStat(pos domain.Position, ovr int) int {
	if pos == domain.PositionGK {
		return ovr + 5
	}
	return 40
}
