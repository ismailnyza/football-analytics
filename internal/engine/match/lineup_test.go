package match

import (
	"testing"

	"github.com/ismael/football-analytics/internal/domain"
)

func TestSelectLineupReturnsElevenPlayers(t *testing.T) {
	assignments, err := SelectLineup(sampleSquad(), "4-3-3", nil)
	if err != nil {
		t.Fatalf("SelectLineup() error = %v", err)
	}
	if len(assignments) != 11 {
		t.Fatalf("len(assignments) = %d, want 11", len(assignments))
	}

	goalkeepers := 0
	for _, assignment := range assignments {
		if assignment.Player.PrimaryPosition == domain.PositionGK {
			goalkeepers++
		}
	}
	if goalkeepers != 1 {
		t.Fatalf("goalkeepers = %d, want 1", goalkeepers)
	}
}

func TestSelectLineupPrefersLockedPlayers(t *testing.T) {
	squad := sampleSquad()

	assignments, err := SelectLineup(squad, "4-3-3", []int64{12})
	if err != nil {
		t.Fatalf("SelectLineup() error = %v", err)
	}

	found := false
	for _, assignment := range assignments {
		if assignment.Player.ID == 12 {
			found = true
		}
	}
	if !found {
		t.Fatal("locked player id 12 missing from lineup")
	}
}

func TestSelectLineupUsesSecondaryPositionWhenNeeded(t *testing.T) {
	squad := sampleSquad()
	squad = append(squad, domain.Player{
		ID:              14,
		FirstName:       "Playmaker",
		LastName:        "Wide",
		PrimaryPosition: domain.PositionCM,
		SecondaryPositions: []domain.Position{
			domain.PositionRW,
		},
		Attributes: domain.PlayerAttributes{Overall: 85},
	})

	assignments, err := SelectLineup(squad, "4-3-3", []int64{14})
	if err != nil {
		t.Fatalf("SelectLineup() error = %v", err)
	}

	for _, assignment := range assignments {
		if assignment.Player.ID == 14 {
			if assignment.Slot.Code != "RW" && assignment.Slot.Code != "RCM" && assignment.Slot.Code != "CM" && assignment.Slot.Code != "LCM" {
				t.Fatalf("locked player assigned to unexpected slot %q", assignment.Slot.Code)
			}
			return
		}
	}
	t.Fatal("player id 14 missing from lineup")
}

func TestSelectLineupRejectsUnsupportedFormation(t *testing.T) {
	if _, err := SelectLineup(sampleSquad(), "3-5-2", nil); err == nil {
		t.Fatal("SelectLineup() error = nil, want error")
	}
}

func sampleSquad() []domain.Player {
	return []domain.Player{
		player(1, "Goal", "Keeper", domain.PositionGK, nil, 80),
		player(2, "Right", "Back", domain.PositionRB, nil, 78),
		player(3, "Center", "Back One", domain.PositionCB, nil, 81),
		player(4, "Center", "Back Two", domain.PositionCB, nil, 79),
		player(5, "Left", "Back", domain.PositionLB, nil, 77),
		player(6, "Anchor", "Mid", domain.PositionDM, []domain.Position{domain.PositionCM}, 80),
		player(7, "Box", "Mid", domain.PositionCM, nil, 79),
		player(8, "Creator", "Mid", domain.PositionAM, []domain.Position{domain.PositionCM}, 82),
		player(9, "Right", "Wing", domain.PositionRW, []domain.Position{domain.PositionAM}, 81),
		player(10, "Striker", "One", domain.PositionST, nil, 84),
		player(11, "Left", "Wing", domain.PositionLW, []domain.Position{domain.PositionAM}, 80),
		player(12, "Utility", "Wide", domain.PositionCM, []domain.Position{domain.PositionLW, domain.PositionRW}, 74),
		player(13, "Reserve", "Forward", domain.PositionST, nil, 73),
	}
}

func player(id int64, firstName, lastName string, primary domain.Position, secondary []domain.Position, overall int) domain.Player {
	return domain.Player{
		ID:                 id,
		FirstName:          firstName,
		LastName:           lastName,
		PrimaryPosition:    primary,
		SecondaryPositions: secondary,
		Attributes: domain.PlayerAttributes{
			Overall:   overall,
			Pace:      overall - 5,
			Shooting:  overall - 4,
			Passing:   overall - 6,
			Defending: overall - 7,
			Keeping:   overall,
		},
	}
}
