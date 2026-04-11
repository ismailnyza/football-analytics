// Package transfer provides transfer valuation and contract logic for the
// long-term world simulation.
package transfer

import (
	"github.com/ismael/football-analytics/internal/domain"
)

// MarketValue returns a deterministic estimated transfer fee for a player
// based on their attributes, age, and potential.
func MarketValue(player domain.Player, age int) int64 {
	ovr := player.Attributes.Overall
	pot := player.Attributes.Potential

	// Base value scaled quadratically with OVR.
	base := int64(ovr) * int64(ovr) * 50_000

	// Age multiplier: peak around 24–27, discount for youth and veterans.
	ageMult := ageMultiplier(age)

	// Potential premium: bonus for high upside.
	potBonus := int64(pot-ovr) * 300_000
	if potBonus < 0 {
		potBonus = 0
	}

	return int64(float64(base)*ageMult) + potBonus
}

func ageMultiplier(age int) float64 {
	switch {
	case age <= 18:
		return 0.7
	case age <= 22:
		return 1.0
	case age <= 27:
		return 1.3
	case age <= 30:
		return 1.0
	case age <= 32:
		return 0.6
	default:
		return 0.3
	}
}

// TransferDecision represents an AI-generated transfer action.
type TransferDecision struct {
	PlayerID  int64
	FromClub  int64
	ToClub    int64
	Fee       int64
	DecisionType string // "buy", "sell", "loan_out", "loan_in"
}

// EvaluateTransfer determines whether a buying club should make an offer for
// a player given their budget and squad need. Returns a decision with type
// "buy" or "pass".
func EvaluateTransfer(
	player domain.Player,
	playerAge int,
	buyingClubBudget int64,
	squadAverageOvr int,
) TransferDecision {
	value := MarketValue(player, playerAge)

	if value > buyingClubBudget {
		return TransferDecision{PlayerID: player.ID, DecisionType: "pass"}
	}

	// Only buy if the player upgrades squad quality.
	if player.Attributes.Overall <= squadAverageOvr {
		return TransferDecision{PlayerID: player.ID, DecisionType: "pass"}
	}

	return TransferDecision{
		PlayerID:     player.ID,
		Fee:          value,
		DecisionType: "buy",
	}
}
