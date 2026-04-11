package match

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ismael/football-analytics/internal/domain"
)

// Slot defines a formation slot and the positions that best fit it.
type Slot struct {
	Code    string
	Allowed []domain.Position
}

// Assignment represents one player's slot within a selected lineup.
type Assignment struct {
	Slot   Slot
	Player domain.Player
}

var formations = map[string][]Slot{
	"4-3-3": {
		{Code: "GK", Allowed: []domain.Position{domain.PositionGK}},
		{Code: "RB", Allowed: []domain.Position{domain.PositionRB}},
		{Code: "RCB", Allowed: []domain.Position{domain.PositionCB}},
		{Code: "LCB", Allowed: []domain.Position{domain.PositionCB}},
		{Code: "LB", Allowed: []domain.Position{domain.PositionLB}},
		{Code: "RCM", Allowed: []domain.Position{domain.PositionCM, domain.PositionDM}},
		{Code: "CM", Allowed: []domain.Position{domain.PositionCM, domain.PositionAM, domain.PositionDM}},
		{Code: "LCM", Allowed: []domain.Position{domain.PositionCM, domain.PositionAM}},
		{Code: "RW", Allowed: []domain.Position{domain.PositionRW, domain.PositionAM}},
		{Code: "ST", Allowed: []domain.Position{domain.PositionST}},
		{Code: "LW", Allowed: []domain.Position{domain.PositionLW, domain.PositionAM}},
	},
	"4-2-3-1": {
		{Code: "GK", Allowed: []domain.Position{domain.PositionGK}},
		{Code: "RB", Allowed: []domain.Position{domain.PositionRB}},
		{Code: "RCB", Allowed: []domain.Position{domain.PositionCB}},
		{Code: "LCB", Allowed: []domain.Position{domain.PositionCB}},
		{Code: "LB", Allowed: []domain.Position{domain.PositionLB}},
		{Code: "RDM", Allowed: []domain.Position{domain.PositionDM, domain.PositionCM}},
		{Code: "LDM", Allowed: []domain.Position{domain.PositionDM, domain.PositionCM}},
		{Code: "RAM", Allowed: []domain.Position{domain.PositionRW, domain.PositionAM}},
		{Code: "CAM", Allowed: []domain.Position{domain.PositionAM, domain.PositionCM}},
		{Code: "LAM", Allowed: []domain.Position{domain.PositionLW, domain.PositionAM}},
		{Code: "ST", Allowed: []domain.Position{domain.PositionST}},
	},
	"4-4-2": {
		{Code: "GK", Allowed: []domain.Position{domain.PositionGK}},
		{Code: "RB", Allowed: []domain.Position{domain.PositionRB}},
		{Code: "RCB", Allowed: []domain.Position{domain.PositionCB}},
		{Code: "LCB", Allowed: []domain.Position{domain.PositionCB}},
		{Code: "LB", Allowed: []domain.Position{domain.PositionLB}},
		{Code: "RM", Allowed: []domain.Position{domain.PositionRW, domain.PositionAM}},
		{Code: "RCM", Allowed: []domain.Position{domain.PositionCM, domain.PositionDM}},
		{Code: "LCM", Allowed: []domain.Position{domain.PositionCM, domain.PositionDM}},
		{Code: "LM", Allowed: []domain.Position{domain.PositionLW, domain.PositionAM}},
		{Code: "RST", Allowed: []domain.Position{domain.PositionST}},
		{Code: "LST", Allowed: []domain.Position{domain.PositionST}},
	},
}

// SelectLineup assigns the best deterministic XI for the requested formation.
func SelectLineup(squad []domain.Player, formation string, lockedIDs []int64) ([]Assignment, error) {
	slots, ok := formations[strings.TrimSpace(formation)]
	if !ok {
		return nil, fmt.Errorf("unsupported formation %q", formation)
	}

	lockedSet := make(map[int64]struct{}, len(lockedIDs))
	for _, id := range lockedIDs {
		lockedSet[id] = struct{}{}
	}

	remaining := append([]domain.Player(nil), squad...)
	slices.SortFunc(remaining, comparePlayers)

	assignments := make([]Assignment, len(slots))
	used := make(map[int64]struct{}, len(slots))
	filledSlots := make(map[int]struct{}, len(slots))

	for _, lockedID := range lockedIDs {
		player, ok := findPlayerByID(remaining, lockedID)
		if !ok {
			return nil, fmt.Errorf("locked player %d not found in squad", lockedID)
		}

		bestSlot := -1
		bestScore := -1
		for idx, slot := range slots {
			if _, alreadyFilled := filledSlots[idx]; alreadyFilled {
				continue
			}
			score := scorePlayerForSlot(player, slot)
			if score < 0 {
				continue
			}
			if bestSlot == -1 || score > bestScore {
				bestSlot = idx
				bestScore = score
			}
		}
		if bestSlot == -1 {
			return nil, fmt.Errorf("locked player %d could not be assigned in formation %s", lockedID, formation)
		}

		assignments[bestSlot] = Assignment{Slot: slots[bestSlot], Player: player}
		used[player.ID] = struct{}{}
		filledSlots[bestSlot] = struct{}{}
	}

	for idx, slot := range slots {
		if _, alreadyFilled := filledSlots[idx]; alreadyFilled {
			continue
		}

		candidateIndex := -1
		candidateScore := -1
		for idx, player := range remaining {
			if _, ok := used[player.ID]; ok {
				continue
			}

			score := scorePlayerForSlot(player, slot)
			if score < 0 {
				continue
			}

			if candidateIndex == -1 || score > candidateScore || (score == candidateScore && comparePlayers(player, remaining[candidateIndex]) < 0) {
				candidateIndex = idx
				candidateScore = score
			}
		}

		if candidateIndex == -1 {
			return nil, fmt.Errorf("could not fill slot %s for formation %s", slot.Code, formation)
		}

		player := remaining[candidateIndex]
		used[player.ID] = struct{}{}
		assignments[idx] = Assignment{Slot: slot, Player: player}
	}

	return assignments, nil
}

func scorePlayerForSlot(player domain.Player, slot Slot) int {
	if isGoalkeeperSlot(slot) {
		if player.PrimaryPosition != domain.PositionGK {
			return -1
		}
		return 500 + player.Attributes.Keeping + player.Attributes.Overall
	}

	if player.PrimaryPosition == domain.PositionGK {
		return -1
	}

	score := player.Attributes.Overall
	for idx, allowed := range slot.Allowed {
		if player.PrimaryPosition == allowed {
			return 400 - idx*10 + score
		}
	}
	for _, secondary := range player.SecondaryPositions {
		for idx, allowed := range slot.Allowed {
			if secondary == allowed {
				return 300 - idx*10 + score
			}
		}
	}
	return 100 + score
}

func isGoalkeeperSlot(slot Slot) bool {
	return len(slot.Allowed) == 1 && slot.Allowed[0] == domain.PositionGK
}

func comparePlayers(left, right domain.Player) int {
	if left.Attributes.Overall != right.Attributes.Overall {
		if left.Attributes.Overall > right.Attributes.Overall {
			return -1
		}
		return 1
	}
	if left.DisplayName() != right.DisplayName() {
		if left.DisplayName() < right.DisplayName() {
			return -1
		}
		return 1
	}
	if left.ID < right.ID {
		return -1
	}
	if left.ID > right.ID {
		return 1
	}
	return 0
}

func findPlayerByID(players []domain.Player, id int64) (domain.Player, bool) {
	for _, player := range players {
		if player.ID == id {
			return player, true
		}
	}
	return domain.Player{}, false
}
