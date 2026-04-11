package app

import (
	"context"
	"fmt"

	"github.com/ismael/football-analytics/internal/domain"
)

// BranchState is the storage-backed club and squad snapshot for the active branch.
type BranchState struct {
	Clubs  []domain.Club
	Squads map[int64][]domain.Player
}

// LoadActiveBranchState loads all clubs and club squads for the configured active branch.
func LoadActiveBranchState(ctx context.Context, cfg Config) (BranchState, error) {
	if cfg.Repo == nil {
		return BranchState{}, fmt.Errorf("repository not configured")
	}
	if cfg.CurrentBranchID() == 0 {
		return BranchState{}, fmt.Errorf("active branch not configured")
	}

	clubs, err := cfg.Repo.ListClubsByBranch(ctx, cfg.CurrentBranchID())
	if err != nil {
		return BranchState{}, fmt.Errorf("list clubs: %w", err)
	}

	squads := make(map[int64][]domain.Player, len(clubs))
	for _, club := range clubs {
		players, err := cfg.Repo.ListPlayersByClub(ctx, cfg.CurrentBranchID(), club.ID)
		if err != nil {
			return BranchState{}, fmt.Errorf("list players for club %d: %w", club.ID, err)
		}
		squads[club.ID] = players
	}

	return BranchState{
		Clubs:  clubs,
		Squads: squads,
	}, nil
}

// EligibleClubs returns clubs whose squads meet a minimum size.
func (s BranchState) EligibleClubs(minPlayers int) []domain.Club {
	out := make([]domain.Club, 0, len(s.Clubs))
	for _, club := range s.Clubs {
		if len(s.Squads[club.ID]) >= minPlayers {
			out = append(out, club)
		}
	}
	return out
}
