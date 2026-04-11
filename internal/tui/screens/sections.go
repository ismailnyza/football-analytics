package screens

import (
	"context"
	"fmt"
	"time"

	"github.com/ismael/football-analytics/internal/app"
	"github.com/ismael/football-analytics/internal/domain"
	"github.com/ismael/football-analytics/internal/tui/screens/club"
	"github.com/ismael/football-analytics/internal/tui/screens/dashboard"
	"github.com/ismael/football-analytics/internal/tui/screens/dataimport"
	"github.com/ismael/football-analytics/internal/tui/screens/matchlab"
	"github.com/ismael/football-analytics/internal/tui/screens/playerdetail"
	"github.com/ismael/football-analytics/internal/tui/screens/scenarios"
	"github.com/ismael/football-analytics/internal/tui/screens/seasonlab"
	"github.com/ismael/football-analytics/internal/tui/screens/squad"
	"github.com/ismael/football-analytics/internal/tui/screens/transfers"
	"github.com/ismael/football-analytics/internal/tui/screens/world"
)

var copyBySection = map[string]string{
	"Settings": "Runtime configuration, display options, and simulation defaults.",
}

func MainPlaceholder(section string) string {
	if copy, ok := copyBySection[section]; ok {
		return fmt.Sprintf("%s\n\nStatus: planned shell section", copy)
	}
	return ShellIdleContent(section)
}

// ContentFor returns placeholder copy for screens that are not yet interactive.
func ContentFor(section string) string {
	return MainPlaceholder(section)
}

func ShellIdleContent(selectedSection string) string {
	return fmt.Sprintf(
		"Highlight: %s\n\n"+
			"The menu is in the left column. If you only see this terminal’s wallpaper,\n"+
			"look at the top-left: Navigation + Main Content sit there (not centered).\n\n"+
			"j / down — move highlight\n"+
			"k / up   — move highlight\n"+
			"enter    — open section (Dashboard, Data / Import, …)\n"+
			"tab      — switch nav vs main focus\n"+
			"esc      — back to nav\n"+
			"?        — key help under status\n"+
			"q        — quit\n",
		selectedSection,
	)
}

// NewScreenFor returns a live interactive Screen for sections that have one,
// or nil for sections that still render as static placeholder text.
func NewScreenFor(section string, cfg app.Config) Screen {
	switch section {
	case "Dashboard":
		m := newDashboard(cfg)
		return &dashboardAdapter{model: m}
	case "Match Lab":
		m := matchlab.NewWithConfig(cfg)
		return &matchlabAdapter{model: m}
	case "Season Lab":
		m := seasonlab.NewWithConfig(cfg)
		return &seasonlabAdapter{model: m}
	case "Club":
		m := newClub(cfg)
		return &clubAdapter{model: m}
	case "Squad":
		m := newSquad(cfg)
		return &squadAdapter{model: m}
	case "Player Detail":
		m := newPlayerDetail(cfg)
		return &playerDetailAdapter{model: m}
	case "Transfers":
		m := transfers.New()
		return &transfersAdapter{model: m}
	case "World":
		m := newWorld(cfg)
		return &worldAdapter{model: m}
	case "Scenarios":
		m := scenarios.NewWithConfig(cfg)
		return &scenariosAdapter{model: m}
	case "Data / Import":
		m := dataimport.New(cfg)
		return &dataimportAdapter{model: m}
	}
	return nil
}

func newDashboard(cfg app.Config) dashboard.Model {
	if cfg.Repo == nil || cfg.CurrentBranchID() == 0 {
		return dashboard.New()
	}

	ctx := context.Background()
	clubs, err := cfg.Repo.ListClubsByBranch(ctx, cfg.CurrentBranchID())
	if err != nil {
		return dashboard.New()
	}

	summary := dashboard.WorldSummary{
		WorldName:     fallback(cfg.CurrentWorldName(), "Local Career"),
		Branch:        fallback(cfg.CurrentBranchName(), "main"),
		Season:        fmt.Sprintf("%d/%02d", time.Now().Year(), (time.Now().Year()+1)%100),
		MatchesPlayed: 0,
		MatchesTotal:  0,
		LeaderName:    "—",
		LeaderPoints:  0,
		Clubs:         len(clubs),
	}
	if len(clubs) > 0 {
		summary.LeaderName = clubs[0].Name
	}
	return dashboard.NewWithSummary(summary)
}

func newClub(cfg app.Config) club.Model {
	if cfg.Repo == nil || cfg.CurrentBranchID() == 0 || cfg.CurrentClubID() == 0 {
		return club.New()
	}

	ctx := context.Background()
	clubs, err := cfg.Repo.ListClubsByBranch(ctx, cfg.CurrentBranchID())
	if err != nil {
		return club.New()
	}
	activeClub, ok := findClub(clubs, cfg.CurrentClubID())
	if !ok {
		return club.New()
	}

	players, err := cfg.Repo.ListPlayersByClub(ctx, cfg.CurrentBranchID(), cfg.CurrentClubID())
	if err != nil {
		return club.New()
	}

	return club.NewWithProfile(club.ClubProfile{
		Club:       activeClub,
		SquadSize:  len(players),
		AverageOvr: averageOverall(players),
		Budget:     int64(50_000_000 + len(players)*2_500_000),
	})
}

func newSquad(cfg app.Config) squad.Model {
	players, ok := loadActivePlayers(cfg)
	if !ok {
		return squad.New()
	}
	return squad.NewWithPlayers(players)
}

func newPlayerDetail(cfg app.Config) playerdetail.Model {
	players, ok := loadActivePlayers(cfg)
	if !ok {
		return playerdetail.New()
	}
	return playerdetail.NewWithPlayers(players)
}

func newWorld(cfg app.Config) world.Model {
	if cfg.Repo == nil || cfg.CurrentWorldID() == 0 {
		return world.New()
	}

	branches, err := cfg.Repo.ListBranches(context.Background(), cfg.CurrentWorldID())
	if err != nil {
		return world.New()
	}

	leader := fallback(cfg.CurrentClubName(), "—")
	return world.NewWithData(
		fallback(cfg.CurrentWorldName(), "Local Career"),
		len(branches),
		time.Now().Year(),
		[]world.LeagueSnapshot{
			{
				Name:   "Local League",
				Leader: leader,
				Points: 0,
				Season: fmt.Sprintf("%d/%02d", time.Now().Year(), (time.Now().Year()+1)%100),
			},
		},
	)
}

func loadActivePlayers(cfg app.Config) ([]domain.Player, bool) {
	if cfg.Repo == nil || cfg.CurrentBranchID() == 0 || cfg.CurrentClubID() == 0 {
		return nil, false
	}
	players, err := cfg.Repo.ListPlayersByClub(context.Background(), cfg.CurrentBranchID(), cfg.CurrentClubID())
	if err != nil || len(players) == 0 {
		return nil, false
	}
	return players, true
}

func findClub(clubs []domain.Club, clubID int64) (domain.Club, bool) {
	for _, current := range clubs {
		if current.ID == clubID {
			return current, true
		}
	}
	return domain.Club{}, false
}

func averageOverall(players []domain.Player) int {
	if len(players) == 0 {
		return 0
	}
	total := 0
	for _, player := range players {
		total += player.Attributes.Overall
	}
	return total / len(players)
}

func fallback(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
