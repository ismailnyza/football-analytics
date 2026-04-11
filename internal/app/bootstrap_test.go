package app

import (
	"context"
	"testing"
)

func TestOpenLocalStoreSeedsDefaultWorld(t *testing.T) {
	cfg := Config{StateDir: t.TempDir()}

	cfg, closeStore, err := OpenLocalStore(context.Background(), cfg)
	if err != nil {
		t.Fatalf("OpenLocalStore() error = %v", err)
	}
	defer func() {
		if err := closeStore(); err != nil {
			t.Fatalf("closeStore() error = %v", err)
		}
	}()

	if cfg.Repo == nil {
		t.Fatal("Repo = nil, want seeded repository")
	}
	if cfg.ActiveWorldName == "" || cfg.ActiveBranchName == "" || cfg.ActiveClubName == "" {
		t.Fatalf("active names not fully populated: %#v", cfg)
	}

	worlds, err := cfg.Repo.ListWorlds(context.Background())
	if err != nil {
		t.Fatalf("ListWorlds() error = %v", err)
	}
	if len(worlds) != 1 {
		t.Fatalf("ListWorlds() len = %d, want 1", len(worlds))
	}

	clubs, err := cfg.Repo.ListClubsByBranch(context.Background(), cfg.ActiveBranchID)
	if err != nil {
		t.Fatalf("ListClubsByBranch() error = %v", err)
	}
	if len(clubs) < 3 {
		t.Fatalf("ListClubsByBranch() len = %d, want at least 3", len(clubs))
	}

	players, err := cfg.Repo.ListPlayersByClub(context.Background(), cfg.ActiveBranchID, cfg.ActiveClubID)
	if err != nil {
		t.Fatalf("ListPlayersByClub() error = %v", err)
	}
	if len(players) < 11 {
		t.Fatalf("ListPlayersByClub() len = %d, want at least 11", len(players))
	}
}
