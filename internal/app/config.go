package app

import "github.com/ismael/football-analytics/internal/storage"

type RuntimeState struct {
	ActiveWorldID    int64
	ActiveWorldName  string
	ActiveBranchID   int64
	ActiveBranchName string
	ActiveClubID     int64
	ActiveClubName   string
}

// Config holds process-wide bootstrap configuration.
type Config struct {
	Name             string
	Version          string
	StateDir         string
	Repo             storage.Repository
	State            *RuntimeState
	ActiveWorldID    int64
	ActiveWorldName  string
	ActiveBranchID   int64
	ActiveBranchName string
	ActiveClubID     int64
	ActiveClubName   string
}

func (c Config) CurrentWorldID() int64 {
	if c.State != nil && c.State.ActiveWorldID != 0 {
		return c.State.ActiveWorldID
	}
	return c.ActiveWorldID
}

func (c Config) CurrentWorldName() string {
	if c.State != nil && c.State.ActiveWorldName != "" {
		return c.State.ActiveWorldName
	}
	return c.ActiveWorldName
}

func (c Config) CurrentBranchID() int64 {
	if c.State != nil && c.State.ActiveBranchID != 0 {
		return c.State.ActiveBranchID
	}
	return c.ActiveBranchID
}

func (c Config) CurrentBranchName() string {
	if c.State != nil && c.State.ActiveBranchName != "" {
		return c.State.ActiveBranchName
	}
	return c.ActiveBranchName
}

func (c Config) CurrentClubID() int64 {
	if c.State != nil && c.State.ActiveClubID != 0 {
		return c.State.ActiveClubID
	}
	return c.ActiveClubID
}

func (c Config) CurrentClubName() string {
	if c.State != nil && c.State.ActiveClubName != "" {
		return c.State.ActiveClubName
	}
	return c.ActiveClubName
}

// DefaultConfig returns the baseline runtime metadata for local builds.
func DefaultConfig() Config {
	return Config{
		Name:    "Football Simulation Engine",
		Version: "0.1.0-bootstrap",
	}
}
