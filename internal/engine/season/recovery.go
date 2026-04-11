package season

import (
	"github.com/ismael/football-analytics/internal/engine/match"
)

// RecoveryParams configures fatigue and injury recovery between matches.
type RecoveryParams struct {
	// DaysRest is the number of days between matches.
	DaysRest int
	// FatigueRecoveryPerDay is the fraction of accumulated fatigue recovered per rest day.
	FatigueRecoveryPerDay float64
}

// DefaultRecoveryParams returns standard week-to-week recovery settings.
func DefaultRecoveryParams() RecoveryParams {
	return RecoveryParams{
		DaysRest:              7,
		FatigueRecoveryPerDay: 0.25,
	}
}

// ApplyFatigueRecovery reduces fatigue values by the configured per-day rate.
// Returns a new map with recovered values (minimum 0).
func ApplyFatigueRecovery(fatigue map[int64]float64, params RecoveryParams) map[int64]float64 {
	recovered := make(map[int64]float64, len(fatigue))
	reduction := params.FatigueRecoveryPerDay * float64(params.DaysRest)
	for playerID, load := range fatigue {
		updated := load - reduction
		if updated < 0 {
			updated = 0
		}
		recovered[playerID] = updated
	}
	return recovered
}

// InjuryRecoveryDays returns the minimum rest days required before a player is
// fit to participate again based on injury severity.
func InjuryRecoveryDays(severity string) int {
	switch severity {
	case "major":
		return 42
	case "moderate":
		return 21
	default: // minor
		return 7
	}
}

// FilterActiveInjuries removes injuries that have healed given the number of
// days elapsed since the match in which they occurred.
func FilterActiveInjuries(injuries []match.Injury, daysElapsed int) []match.Injury {
	active := make([]match.Injury, 0, len(injuries))
	for _, inj := range injuries {
		if daysElapsed < InjuryRecoveryDays(inj.Severity) {
			active = append(active, inj)
		}
	}
	return active
}
