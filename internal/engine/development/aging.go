package development

import (
	"time"

	"github.com/ismael/football-analytics/internal/domain"
)

// AgeAtDate computes a player's age on a given date using their date of birth.
// Returns 0 if the date of birth is zero.
func AgeAtDate(player domain.Player, on time.Time) int {
	if player.DateOfBirth.IsZero() {
		return 0
	}
	dob := player.DateOfBirth
	age := on.Year() - dob.Year()
	// Adjust if birthday hasn't occurred yet this year.
	birthday := time.Date(on.Year(), dob.Month(), dob.Day(), 0, 0, 0, 0, on.Location())
	if on.Before(birthday) {
		age--
	}
	return age
}

// AgingResult holds the outcome of processing one year of aging for a squad.
type AgingResult struct {
	Updated  []domain.Player // players with modified attributes
	Retiring []domain.Player // players flagged for retirement
}

// ProcessSquadAging applies one simulation year of development and retirement
// checks to a squad. The reference date is used to compute current ages.
func ProcessSquadAging(squad []domain.Player, refDate time.Time, params DevelopmentParams) AgingResult {
	result := AgingResult{
		Updated:  make([]domain.Player, 0, len(squad)),
		Retiring: make([]domain.Player, 0),
	}
	for _, player := range squad {
		age := AgeAtDate(player, refDate)
		if age == 0 {
			// No date of birth — skip development but include in squad.
			result.Updated = append(result.Updated, player)
			continue
		}
		if ShouldRetire(age) {
			result.Retiring = append(result.Retiring, player)
			continue
		}
		result.Updated = append(result.Updated, AgePlayer(player, age, params))
	}
	return result
}
