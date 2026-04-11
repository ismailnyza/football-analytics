// Package development provides the player development engine — age-based
// attribute progression, peak years, and decline modelling.
package development

import (
	"github.com/ismael/football-analytics/internal/domain"
)

const (
	peakAgeMin = 26
	peakAgeMax = 29
	youthAge   = 17
	retireAge  = 35
)

// DevelopmentParams configures the annual progression model.
type DevelopmentParams struct {
	// GrowthScale is the base attribute gain per year below peak (0.0–1.0).
	GrowthScale float64
	// DeclineScale is the base attribute loss per year above peak (0.0–1.0).
	DeclineScale float64
}

// DefaultDevelopmentParams returns sensible defaults.
func DefaultDevelopmentParams() DevelopmentParams {
	return DevelopmentParams{
		GrowthScale:  0.8,
		DeclineScale: 0.6,
	}
}

// AgePlayer advances a player's attributes by one simulation year.
// Attributes grow toward potential before the peak window, hold during it,
// and decline after it. Returns the updated player.
func AgePlayer(player domain.Player, currentAge int, params DevelopmentParams) domain.Player {
	updated := player

	switch {
	case currentAge < peakAgeMin:
		// Growth phase: move OVR toward potential.
		gap := player.Attributes.Potential - player.Attributes.Overall
		if gap > 0 {
			gain := int(float64(gap) * params.GrowthScale * growthRate(currentAge))
			if gain < 1 {
				gain = 1
			}
			updated.Attributes.Overall = min(player.Attributes.Overall+gain, player.Attributes.Potential)
			updated.Attributes = growAttributes(player.Attributes, gain)
		}

	case currentAge >= peakAgeMin && currentAge <= peakAgeMax:
		// Peak: no change.

	default:
		// Decline phase.
		yearsOver := currentAge - peakAgeMax
		loss := int(float64(yearsOver) * params.DeclineScale * 2)
		if loss < 1 {
			loss = 1
		}
		updated.Attributes.Overall = max(player.Attributes.Overall-loss, 40)
		updated.Attributes = declineAttributes(player.Attributes, loss)
	}

	return updated
}

func growthRate(age int) float64 {
	// Faster growth in late teens / early 20s, slowing toward peak.
	switch {
	case age <= 19:
		return 1.0
	case age <= 22:
		return 0.7
	case age <= 24:
		return 0.4
	default:
		return 0.2
	}
}

func growAttributes(attrs domain.PlayerAttributes, gain int) domain.PlayerAttributes {
	half := gain / 2
	if half < 1 {
		half = 1
	}
	attrs.Overall = clamp(attrs.Overall+gain, 40, 99)
	attrs.Pace = clamp(attrs.Pace+half, 40, 99)
	attrs.Shooting = clamp(attrs.Shooting+half, 40, 99)
	attrs.Passing = clamp(attrs.Passing+gain, 40, 99)
	attrs.Defending = clamp(attrs.Defending+half, 40, 99)
	return attrs
}

func declineAttributes(attrs domain.PlayerAttributes, loss int) domain.PlayerAttributes {
	attrs.Overall = clamp(attrs.Overall-loss, 40, 99)
	attrs.Pace = clamp(attrs.Pace-loss, 40, 99)
	attrs.Shooting = clamp(attrs.Shooting-(loss/2), 40, 99)
	attrs.Passing = clamp(attrs.Passing-(loss/2), 40, 99)
	attrs.Defending = clamp(attrs.Defending-(loss/2), 40, 99)
	return attrs
}

// ShouldRetire reports whether a player should retire this year.
func ShouldRetire(currentAge int) bool {
	return currentAge >= retireAge
}

// IsYouth reports whether a player is in the youth development bracket.
func IsYouth(currentAge int) bool {
	return currentAge <= youthAge+3
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
