package match

import (
	"fmt"

	"github.com/ismael/football-analytics/internal/domain"
)

const (
	// DefaultTickCount is the baseline V1 match length.
	DefaultTickCount = 900
	secondsPerTick   = 6
)

// TeamPlan is the pre-match data needed by the tick loop.
type TeamPlan struct {
	Club    domain.Club
	Lineup  []Assignment
	Attack  int
	Control int
	Defence int
}

// State is the deterministic in-memory match state across ticks.
type State struct {
	Seed       int64
	Tick       int
	HomeGoals  int
	AwayGoals  int
	Possession string
	Events     []domain.MatchEvent
}

// Summary is the terminal output of the tick loop.
type Summary struct {
	TotalTicks int
	HomeGoals  int
	AwayGoals  int
	Events     []domain.MatchEvent
}

// BuildTeamPlan derives deterministic team strengths from the selected lineup.
func BuildTeamPlan(club domain.Club, lineup []Assignment) TeamPlan {
	plan := TeamPlan{Club: club, Lineup: lineup}
	for _, assignment := range lineup {
		player := assignment.Player
		plan.Attack += player.Attributes.Shooting + player.Attributes.Pace
		plan.Control += player.Attributes.Passing + player.Attributes.Overall
		plan.Defence += player.Attributes.Defending
		if player.PrimaryPosition == domain.PositionGK {
			plan.Defence += player.Attributes.Keeping * 2
		}
	}
	if len(lineup) > 0 {
		plan.Attack /= len(lineup)
		plan.Control /= len(lineup)
		plan.Defence /= len(lineup)
	}
	return plan
}

// RunTickLoop executes a deterministic V1 match loop.
func RunTickLoop(seed int64, home, away TeamPlan, ticks int) Summary {
	if ticks <= 0 {
		ticks = DefaultTickCount
	}

	state := State{
		Seed:       seed,
		Possession: "home",
		Events:     make([]domain.MatchEvent, 0, ticks/6),
	}

	for tick := 1; tick <= ticks; tick++ {
		state.Tick = tick
		state.Possession = possessionForTick(seed, tick, home, away)

		attacking := home
		defending := away
		if state.Possession == "away" {
			attacking = away
			defending = home
		}

		if shouldCreateChance(seed, tick, attacking, defending) {
			minute := tickToMinute(tick)
			eventType := "chance"
			if shouldScore(seed, tick, attacking, defending) {
				eventType = "goal"
				if state.Possession == "home" {
					state.HomeGoals++
				} else {
					state.AwayGoals++
				}
			}
			state.Events = append(state.Events, domain.MatchEvent{
				Tick:      tick,
				Minute:    minute,
				Type:      eventType,
				PitchZone: attackZone(attacking),
				PayloadRaw: fmt.Sprintf(
					`{"team":"%s","home_goals":%d,"away_goals":%d}`,
					attacking.Club.ShortName,
					state.HomeGoals,
					state.AwayGoals,
				),
			})
		}
	}

	return Summary{
		TotalTicks: ticks,
		HomeGoals:  state.HomeGoals,
		AwayGoals:  state.AwayGoals,
		Events:     state.Events,
	}
}

func possessionForTick(seed int64, tick int, home, away TeamPlan) string {
	homeBias := home.Control - away.Control
	value := int(seed%17) + tick + homeBias
	if value%5 == 0 || value%5 == 1 {
		return "away"
	}
	return "home"
}

func shouldCreateChance(seed int64, tick int, attacking, defending TeamPlan) bool {
	swing := attacking.Attack + attacking.Control - defending.Defence
	window := (tick + int(seed%13) + max(swing, -20)) % 11
	return window == 0 || window == 1
}

func shouldScore(seed int64, tick int, attacking, defending TeamPlan) bool {
	finishing := attacking.Attack - defending.Defence/2
	window := (tick*3 + int(seed%29) + max(finishing, -15)) % 7
	return window == 0
}

func tickToMinute(tick int) int {
	return ((tick - 1) * secondsPerTick) / 60
}

func attackZone(plan TeamPlan) string {
	if plan.Attack >= plan.Control {
		return "final-third"
	}
	return "half-space"
}
