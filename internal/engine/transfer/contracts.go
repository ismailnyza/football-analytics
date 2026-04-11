package transfer

import (
	"github.com/ismael/football-analytics/internal/domain"
)

// Contract represents a player's employment agreement with a club.
type Contract struct {
	PlayerID      int64
	ClubID        int64
	WeeklyWage    int64 // in currency units
	YearsRemaining int
	ExpiryYear    int
}

// ContractStatus categorises the renewal urgency of a contract.
type ContractStatus string

const (
	StatusSecure  ContractStatus = "secure"   // 2+ years
	StatusMonitor ContractStatus = "monitor"  // 1 year
	StatusExpiring ContractStatus = "expiring" // final year / expired
)

// ContractFor derives a plausible contract for a player based on their OVR
// and the current simulation year.
func ContractFor(player domain.Player, clubID int64, currentYear, yearsRemaining int) Contract {
	weekly := weeklyWage(player.Attributes.Overall)
	return Contract{
		PlayerID:       player.ID,
		ClubID:         clubID,
		WeeklyWage:     weekly,
		YearsRemaining: yearsRemaining,
		ExpiryYear:     currentYear + yearsRemaining,
	}
}

// Status returns the renewal urgency for a contract.
func (c Contract) Status() ContractStatus {
	switch {
	case c.YearsRemaining >= 2:
		return StatusSecure
	case c.YearsRemaining == 1:
		return StatusMonitor
	default:
		return StatusExpiring
	}
}

// RenewalOffer returns a new contract offer for the player with an optional
// wage increase percentage (e.g. 0.10 = 10% raise).
func RenewalOffer(existing Contract, wageIncrease float64, newYears int) Contract {
	newWage := int64(float64(existing.WeeklyWage) * (1.0 + wageIncrease))
	return Contract{
		PlayerID:       existing.PlayerID,
		ClubID:         existing.ClubID,
		WeeklyWage:     newWage,
		YearsRemaining: newYears,
		ExpiryYear:     existing.ExpiryYear - existing.YearsRemaining + newYears,
	}
}

// AnnualWageBill computes the total yearly wage cost for a set of contracts.
func AnnualWageBill(contracts []Contract) int64 {
	var total int64
	for _, c := range contracts {
		total += c.WeeklyWage * 52
	}
	return total
}

// weeklyWage returns a plausible weekly wage based on overall rating.
func weeklyWage(ovr int) int64 {
	// Rough scale: OVR 99 ≈ £350k/week, OVR 60 ≈ £10k/week.
	if ovr < 60 {
		ovr = 60
	}
	if ovr > 99 {
		ovr = 99
	}
	// Exponential-ish mapping.
	base := int64(ovr-59) * int64(ovr-59) * 400
	return base
}

// ExpiringContracts returns players whose contracts expire this year or next.
func ExpiringContracts(players []domain.Player, contracts map[int64]Contract) []Contract {
	expiring := make([]Contract, 0)
	for _, p := range players {
		c, ok := contracts[p.ID]
		if !ok {
			continue
		}
		if c.Status() != StatusSecure {
			expiring = append(expiring, c)
		}
	}
	return expiring
}
