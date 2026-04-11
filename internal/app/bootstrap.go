package app

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/ismael/football-analytics/internal/domain"
	"github.com/ismael/football-analytics/internal/storage/sqlite"
	_ "modernc.org/sqlite"
)

// OpenLocalStore opens the default SQLite store, applies migrations, and
// ensures a baseline demo world exists for the TUI to browse.
func OpenLocalStore(ctx context.Context, cfg Config) (Config, func() error, error) {
	stateDir, err := cfg.ResolveStateDir()
	if err != nil {
		return cfg, nil, fmt.Errorf("resolve state dir: %w", err)
	}
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		return cfg, nil, fmt.Errorf("create state dir: %w", err)
	}

	dbPath := sqlite.DefaultPath(stateDir)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return cfg, nil, fmt.Errorf("open sqlite db: %w", err)
	}

	closeFn := func() error { return db.Close() }
	store := sqlite.NewStore(db)
	if err := store.Migrate(ctx); err != nil {
		_ = db.Close()
		return cfg, nil, fmt.Errorf("migrate sqlite db %s: %w", dbPath, err)
	}

	worldID, worldName, branchID, branchName, clubID, clubName, err := ensureDemoWorld(ctx, store)
	if err != nil {
		_ = db.Close()
		return cfg, nil, err
	}

	cfg.StateDir = stateDir
	cfg.Repo = store
	cfg.ActiveWorldID = worldID
	cfg.ActiveWorldName = worldName
	cfg.ActiveBranchID = branchID
	cfg.ActiveBranchName = branchName
	cfg.ActiveClubID = clubID
	cfg.ActiveClubName = clubName
	cfg.State = &RuntimeState{
		ActiveWorldID:    worldID,
		ActiveWorldName:  worldName,
		ActiveBranchID:   branchID,
		ActiveBranchName: branchName,
		ActiveClubID:     clubID,
		ActiveClubName:   clubName,
	}
	return cfg, closeFn, nil
}

func ensureDemoWorld(ctx context.Context, store *sqlite.Store) (int64, string, int64, string, int64, string, error) {
	worldID, worldName, err := firstWorld(ctx, store)
	if err != nil {
		return 0, "", 0, "", 0, "", fmt.Errorf("load existing world: %w", err)
	}
	if worldID == 0 {
		world, err := store.CreateWorld(ctx, domain.World{Name: "Local Career", Seed: 1})
		if err != nil {
			return 0, "", 0, "", 0, "", fmt.Errorf("create demo world: %w", err)
		}
		branch, err := store.CreateBranch(ctx, domain.Branch{WorldID: world.ID, Name: "main"})
		if err != nil {
			return 0, "", 0, "", 0, "", fmt.Errorf("create demo branch: %w", err)
		}

		clubs, players := demoWorldRoster(branch.ID)
		for i := range clubs {
			createdClub, err := store.CreateClub(ctx, clubs[i])
			if err != nil {
				return 0, "", 0, "", 0, "", fmt.Errorf("create demo club %q: %w", clubs[i].Name, err)
			}
			for _, player := range players[i] {
				player.BranchID = branch.ID
				player.ClubID = &createdClub.ID
				if _, err := store.CreatePlayer(ctx, player); err != nil {
					return 0, "", 0, "", 0, "", fmt.Errorf("create demo player for %q: %w", createdClub.Name, err)
				}
			}
			clubs[i] = createdClub
		}

		if err := ensureMinimumSquads(ctx, store, branch.ID, clubs); err != nil {
			return 0, "", 0, "", 0, "", err
		}

		return world.ID, world.Name, branch.ID, branch.Name, clubs[0].ID, clubs[0].Name, nil
	}

	branches, err := store.ListBranches(ctx, worldID)
	if err != nil {
		return 0, "", 0, "", 0, "", fmt.Errorf("list branches: %w", err)
	}
	if len(branches) == 0 {
		return 0, "", 0, "", 0, "", fmt.Errorf("world %d has no branches", worldID)
	}
	branch := branches[0]

	clubs, err := store.ListClubsByBranch(ctx, branch.ID)
	if err != nil {
		return 0, "", 0, "", 0, "", fmt.Errorf("list clubs: %w", err)
	}
	if len(clubs) == 0 {
		return 0, "", 0, "", 0, "", fmt.Errorf("branch %d has no clubs", branch.ID)
	}
	if err := ensureMinimumSquads(ctx, store, branch.ID, clubs); err != nil {
		return 0, "", 0, "", 0, "", err
	}

	return worldID, worldName, branch.ID, branch.Name, clubs[0].ID, clubs[0].Name, nil
}

func firstWorld(ctx context.Context, store *sqlite.Store) (int64, string, error) {
	worlds, err := store.ListWorlds(ctx)
	if err != nil {
		return 0, "", err
	}
	if len(worlds) == 0 {
		return 0, "", nil
	}
	return worlds[0].ID, worlds[0].Name, nil
}

func demoWorldRoster(branchID int64) ([]domain.Club, [][]domain.Player) {
	clubs := []domain.Club{
		{Name: "Arsenal FC", ShortName: "ARS", Country: "England", Division: "Premier League", BranchID: branchID},
		{Name: "Liverpool FC", ShortName: "LIV", Country: "England", Division: "Premier League", BranchID: branchID},
		{Name: "Manchester City", ShortName: "MCI", Country: "England", Division: "Premier League", BranchID: branchID},
	}
	return clubs, [][]domain.Player{
		makePlayers("ARS", []playerSeed{
			{"David", "Raya", domain.PositionGK, 83, 86},
			{"Ben", "White", domain.PositionRB, 82, 84},
			{"William", "Saliba", domain.PositionCB, 88, 91},
			{"Gabriel", "Magalhaes", domain.PositionCB, 85, 87},
			{"Oleksandr", "Zinchenko", domain.PositionLB, 80, 82},
			{"Declan", "Rice", domain.PositionDM, 87, 89},
			{"Jorginho", "Frello", domain.PositionCM, 79, 79},
			{"Martin", "Odegaard", domain.PositionAM, 88, 90},
			{"Bukayo", "Saka", domain.PositionRW, 89, 93},
			{"Gabriel", "Martinelli", domain.PositionLW, 85, 88},
			{"Kai", "Havertz", domain.PositionST, 82, 85},
		}),
		makePlayers("LIV", []playerSeed{
			{"Alisson", "Becker", domain.PositionGK, 89, 89},
			{"Trent", "Alexander-Arnold", domain.PositionRB, 87, 89},
			{"Virgil", "van Dijk", domain.PositionCB, 89, 89},
			{"Ibrahima", "Konate", domain.PositionCB, 84, 86},
			{"Andy", "Robertson", domain.PositionLB, 84, 84},
			{"Alexis", "Mac Allister", domain.PositionCM, 85, 87},
			{"Dominik", "Szoboszlai", domain.PositionAM, 84, 87},
			{"Mohamed", "Salah", domain.PositionRW, 90, 90},
			{"Luis", "Diaz", domain.PositionLW, 84, 86},
			{"Darwin", "Nunez", domain.PositionST, 83, 86},
			{"Ryan", "Gravenberch", domain.PositionDM, 82, 86},
		}),
		makePlayers("MCI", []playerSeed{
			{"Ederson", "Moraes", domain.PositionGK, 88, 88},
			{"Kyle", "Walker", domain.PositionRB, 84, 84},
			{"Ruben", "Dias", domain.PositionCB, 89, 90},
			{"Josko", "Gvardiol", domain.PositionLB, 86, 90},
			{"John", "Stones", domain.PositionCB, 86, 87},
			{"Rodri", "", domain.PositionDM, 91, 91},
			{"Bernardo", "Silva", domain.PositionCM, 88, 88},
			{"Kevin", "De Bruyne", domain.PositionAM, 91, 91},
			{"Savinho", "", domain.PositionRW, 80, 86},
			{"Phil", "Foden", domain.PositionLW, 88, 92},
			{"Erling", "Haaland", domain.PositionST, 91, 94},
		}),
	}
}

type playerSeed struct {
	first string
	last  string
	pos   domain.Position
	ovr   int
	pot   int
}

func makePlayers(clubCode string, seeds []playerSeed) []domain.Player {
	players := make([]domain.Player, 0, len(seeds))
	baseDOB := time.Date(1998, time.July, 1, 0, 0, 0, 0, time.UTC)
	for i, seed := range seeds {
		players = append(players, domain.Player{
			ExtKey:          fmt.Sprintf("%s-%02d", clubCode, i+1),
			FirstName:       seed.first,
			LastName:        seed.last,
			DateOfBirth:     baseDOB.AddDate(-(i % 7), i%12, i),
			Nationality:     "Unknown",
			PrimaryPosition: seed.pos,
			Attributes: domain.PlayerAttributes{
				Overall:   seed.ovr,
				Potential: seed.pot,
				Pace:      clamp(seed.ovr-2+i%5, 45, 99),
				Shooting:  clamp(seed.ovr-3+i%6, 40, 99),
				Passing:   clamp(seed.ovr-1+i%4, 40, 99),
				Defending: clamp(seed.ovr-4+i%7, 40, 99),
				Keeping:   keepingFor(seed.pos, seed.ovr),
			},
		})
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

func keepingFor(pos domain.Position, overall int) int {
	if pos == domain.PositionGK {
		return clamp(overall+3, 60, 99)
	}
	return 40
}

func ensureMinimumSquads(ctx context.Context, store *sqlite.Store, branchID int64, clubs []domain.Club) error {
	for _, club := range clubs {
		players, err := store.ListPlayersByClub(ctx, branchID, club.ID)
		if err != nil {
			return fmt.Errorf("load existing squad for %q: %w", club.Name, err)
		}
		if len(players) >= 11 {
			continue
		}
		extras := fillerPlayers(club, len(players), 11-len(players))
		for _, player := range extras {
			player.BranchID = branchID
			player.ClubID = &club.ID
			if _, err := store.CreatePlayer(ctx, player); err != nil {
				return fmt.Errorf("seed filler player for %q: %w", club.Name, err)
			}
		}
	}
	return nil
}

func fillerPlayers(club domain.Club, existingCount, needed int) []domain.Player {
	positions := []domain.Position{
		domain.PositionGK,
		domain.PositionRB,
		domain.PositionCB,
		domain.PositionCB,
		domain.PositionLB,
		domain.PositionDM,
		domain.PositionCM,
		domain.PositionAM,
		domain.PositionRW,
		domain.PositionLW,
		domain.PositionST,
	}
	out := make([]domain.Player, 0, needed)
	for i := 0; i < needed; i++ {
		idx := existingCount + i
		pos := positions[idx%len(positions)]
		overall := 68 + (idx % 7)
		out = append(out, domain.Player{
			ExtKey:          fmt.Sprintf("%s-fill-%02d", club.ShortName, idx+1),
			FirstName:       club.ShortName,
			LastName:        fmt.Sprintf("Prospect %02d", idx+1),
			DateOfBirth:     time.Date(2002, time.January, 1, 0, 0, 0, 0, time.UTC).AddDate(-(idx % 3), idx%12, idx),
			Nationality:     club.Country,
			PrimaryPosition: pos,
			Attributes: domain.PlayerAttributes{
				Overall:   overall,
				Potential: overall + 6,
				Pace:      clamp(overall+1, 40, 99),
				Shooting:  clamp(overall-1, 40, 99),
				Passing:   clamp(overall, 40, 99),
				Defending: clamp(overall-2, 40, 99),
				Keeping:   keepingFor(pos, overall),
			},
		})
	}
	return out
}
