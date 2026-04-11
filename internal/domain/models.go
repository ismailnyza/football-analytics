package domain

import (
	"fmt"
	"strings"
	"time"
)

// Position represents a football position in canonical storage and engine models.
type Position string

const (
	PositionGK Position = "GK"
	PositionRB Position = "RB"
	PositionCB Position = "CB"
	PositionLB Position = "LB"
	PositionDM Position = "DM"
	PositionCM Position = "CM"
	PositionAM Position = "AM"
	PositionRW Position = "RW"
	PositionLW Position = "LW"
	PositionST Position = "ST"
)

var validPositions = map[Position]struct{}{
	PositionGK: {},
	PositionRB: {},
	PositionCB: {},
	PositionLB: {},
	PositionDM: {},
	PositionCM: {},
	PositionAM: {},
	PositionRW: {},
	PositionLW: {},
	PositionST: {},
}

// ParsePosition normalizes a position code into the canonical enum.
func ParsePosition(raw string) (Position, error) {
	position := Position(strings.ToUpper(strings.TrimSpace(raw)))
	if _, ok := validPositions[position]; !ok {
		return "", fmt.Errorf("unknown position %q", raw)
	}
	return position, nil
}

// Valid reports whether a position is part of the canonical enum.
func (p Position) Valid() bool {
	_, ok := validPositions[p]
	return ok
}

// World is the root simulation container.
type World struct {
	ID        int64
	Name      string
	Seed      int64
	CreatedAt time.Time
}

// Branch is an alternate timeline within a world.
type Branch struct {
	ID                 int64
	WorldID            int64
	ParentBranchID     *int64
	Name               string
	CreatedFromMatchID *int64
	CreatedAt          time.Time
}

// IsRoot reports whether the branch is the canonical world timeline.
func (b Branch) IsRoot() bool {
	return b.ParentBranchID == nil
}

// Club stores canonical club identity and world membership.
type Club struct {
	ID        int64
	BranchID  int64
	ExtKey    string
	Name      string
	ShortName string
	Country   string
	Division  string
}

// PlayerAttributes contains stable capability ratings used by the simulation engine.
type PlayerAttributes struct {
	Overall   int
	Potential int
	Pace      int
	Shooting  int
	Passing   int
	Defending int
	Keeping   int
}

// Player stores canonical footballer identity plus baseline simulation attributes.
type Player struct {
	ID                 int64
	BranchID           int64
	ClubID             *int64
	ExtKey             string
	FirstName          string
	LastName           string
	PreferredName      string
	DateOfBirth        time.Time
	Nationality        string
	PrimaryPosition    Position
	SecondaryPositions []Position
	Attributes         PlayerAttributes
}

// DisplayName returns the preferred name when present, otherwise full name.
func (p Player) DisplayName() string {
	if strings.TrimSpace(p.PreferredName) != "" {
		return p.PreferredName
	}
	return strings.TrimSpace(strings.TrimSpace(p.FirstName) + " " + strings.TrimSpace(p.LastName))
}

// Season identifies a branch-specific competition year.
type Season struct {
	ID        int64
	BranchID  int64
	Label     string
	StartYear int
}

// Fixture stores the scheduled pairing before a match is simulated.
type Fixture struct {
	ID         int64
	SeasonID   int64
	Matchday   int
	HomeClubID int64
	AwayClubID int64
	Scheduled  time.Time
}

// Validate checks invariant fixture rules required by the engine and persistence layers.
func (f Fixture) Validate() error {
	if f.SeasonID == 0 {
		return fmt.Errorf("season id is required")
	}
	if f.Matchday <= 0 {
		return fmt.Errorf("matchday must be positive")
	}
	if f.HomeClubID == 0 || f.AwayClubID == 0 {
		return fmt.Errorf("home and away clubs are required")
	}
	if f.HomeClubID == f.AwayClubID {
		return fmt.Errorf("home and away clubs must differ")
	}
	return nil
}

// Match captures the deterministic simulation output for a fixture in a branch.
type Match struct {
	ID          int64
	FixtureID   int64
	BranchID    int64
	Seed        int64
	HomeGoals   int
	AwayGoals   int
	TickCount   int
	SimulatedAt time.Time
	Status      string
}

// MatchEvent captures an event generated within a match tick.
type MatchEvent struct {
	ID         int64
	MatchID    int64
	Tick       int
	Minute     int
	TeamClubID *int64
	PlayerID   *int64
	Type       string
	PitchZone  string
	PayloadRaw string
}
