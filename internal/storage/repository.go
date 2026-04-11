package storage

import (
	"context"
	"errors"

	"github.com/ismael/football-analytics/internal/domain"
)

// ErrNotFound is returned when a repository lookup has no matching row.
var ErrNotFound = errors.New("storage: not found")

// Migrator applies pending schema changes.
type Migrator interface {
	Migrate(ctx context.Context) error
}

// WorldRepository persists worlds and branches.
type WorldRepository interface {
	CreateWorld(ctx context.Context, world domain.World) (domain.World, error)
	ListWorlds(ctx context.Context) ([]domain.World, error)
	CreateBranch(ctx context.Context, branch domain.Branch) (domain.Branch, error)
	ListBranches(ctx context.Context, worldID int64) ([]domain.Branch, error)
}

// ClubRepository persists club identity records.
type ClubRepository interface {
	CreateClub(ctx context.Context, club domain.Club) (domain.Club, error)
	ListClubsByBranch(ctx context.Context, branchID int64) ([]domain.Club, error)
}

// PlayerRepository persists player identity records.
type PlayerRepository interface {
	CreatePlayer(ctx context.Context, player domain.Player) (domain.Player, error)
	ListPlayersByClub(ctx context.Context, branchID, clubID int64) ([]domain.Player, error)
}

// SeasonRepository persists season and fixture records.
type SeasonRepository interface {
	CreateSeason(ctx context.Context, season domain.Season) (domain.Season, error)
	ListSeasonsByBranch(ctx context.Context, branchID int64) ([]domain.Season, error)
	CreateFixture(ctx context.Context, fixture domain.Fixture) (domain.Fixture, error)
	ListFixturesBySeason(ctx context.Context, seasonID int64) ([]domain.Fixture, error)
}

// MatchRepository persists match summaries and event logs.
type MatchRepository interface {
	CreateMatch(ctx context.Context, match domain.Match) (domain.Match, error)
	ListMatchesBySeason(ctx context.Context, seasonID int64) ([]domain.Match, error)
	SaveMatchEvents(ctx context.Context, events []domain.MatchEvent) error
	ListMatchEvents(ctx context.Context, matchID int64) ([]domain.MatchEvent, error)
}

// Repository groups the persistence contracts used by services and the TUI.
type Repository interface {
	Migrator
	WorldRepository
	ClubRepository
	PlayerRepository
	SeasonRepository
	MatchRepository
}
