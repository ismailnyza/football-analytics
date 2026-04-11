package development

import (
	"fmt"
	"time"

	"github.com/ismael/football-analytics/internal/domain"
)

// YouthIntakeParams controls the annual generation of youth players.
type YouthIntakeParams struct {
	// Count is the number of new players generated per intake.
	Count int
	// MinPotential and MaxPotential are the OVR potential range for intakes.
	MinPotential int
	MaxPotential int
	// YouthAge is the age assigned to all intake players.
	YouthAge int
}

// DefaultYouthIntakeParams returns standard intake settings.
func DefaultYouthIntakeParams() YouthIntakeParams {
	return YouthIntakeParams{
		Count:        3,
		MinPotential: 65,
		MaxPotential: 85,
		YouthAge:     17,
	}
}

// positionPool lists the positions used when generating youth players.
var positionPool = []domain.Position{
	domain.PositionGK,
	domain.PositionCB, domain.PositionCB,
	domain.PositionRB, domain.PositionLB,
	domain.PositionDM, domain.PositionCM, domain.PositionCM,
	domain.PositionAM, domain.PositionRW, domain.PositionLW,
	domain.PositionST, domain.PositionST,
}

// GenerateYouthIntake produces a deterministic set of youth players for the
// given club. The seed is used to derive attribute variations reproducibly.
func GenerateYouthIntake(clubID, branchID, seed int64, intakeYear int, params YouthIntakeParams) []domain.Player {
	players := make([]domain.Player, 0, params.Count)
	dob := time.Date(intakeYear-params.YouthAge, time.July, 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < params.Count; i++ {
		// Deterministic variation from seed.
		variation := int((seed*int64(i+1)*31 + int64(i*17)) % int64(params.MaxPotential-params.MinPotential+1))
		potential := params.MinPotential + variation
		overall := potential - 15 - int((seed*int64(i+3))%10)
		if overall < 40 {
			overall = 40
		}

		posIdx := int((seed*int64(i+5)+int64(i*13))%int64(len(positionPool)))
		pos := positionPool[posIdx]

		keeping := 40
		if pos == domain.PositionGK {
			keeping = overall + 5
		}

		players = append(players, domain.Player{
			BranchID:        branchID,
			ClubID:          &clubID,
			FirstName:       fmt.Sprintf("Youth%d", intakeYear),
			LastName:        fmt.Sprintf("Player%d", i+1),
			DateOfBirth:     dob,
			PrimaryPosition: pos,
			Attributes: domain.PlayerAttributes{
				Overall:   overall,
				Potential: potential,
				Pace:      clamp(overall+int((seed*int64(i+2))%8)-4, 40, 99),
				Shooting:  clamp(overall-5+int((seed*int64(i+6))%6), 40, 99),
				Passing:   clamp(overall-3+int((seed*int64(i+4))%7), 40, 99),
				Defending: clamp(overall-6+int((seed*int64(i+7))%8), 40, 99),
				Keeping:   keeping,
			},
		})
	}
	return players
}
