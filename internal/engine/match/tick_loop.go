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
	Seed        int64
	Tick        int
	HomeGoals   int
	AwayGoals   int
	Possession  string
	HomeFatigue map[int64]float64
	AwayFatigue map[int64]float64
	Events      []domain.MatchEvent
}

// Summary is the terminal output of the tick loop.
type Summary struct {
	TotalTicks         int
	HomeGoals          int
	AwayGoals          int
	HomeAverageFatigue float64
	AwayAverageFatigue float64
	Events             []domain.MatchEvent
}

type attackPhase struct {
	zone     string
	advanced bool
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
		Seed:        seed,
		Possession:  "home",
		HomeFatigue: initialFatigue(home.Lineup),
		AwayFatigue: initialFatigue(away.Lineup),
		Events:      make([]domain.MatchEvent, 0, ticks/6),
	}

	for tick := 1; tick <= ticks; tick++ {
		state.Tick = tick
		effectiveHome := fatigueAdjustedPlan(home, state.HomeFatigue)
		effectiveAway := fatigueAdjustedPlan(away, state.AwayFatigue)
		state.Possession = possessionForTick(seed, tick, effectiveHome, effectiveAway)

		attacking := effectiveHome
		defending := effectiveAway
		if state.Possession == "away" {
			attacking = effectiveAway
			defending = effectiveHome
		}

		for _, event := range resolveTickActions(seed, tick, attacking, defending, state.Possession, &state) {
			state.Events = append(state.Events, event)
		}
		applyFatigueTick(state.HomeFatigue, home.Lineup, state.Possession == "home")
		applyFatigueTick(state.AwayFatigue, away.Lineup, state.Possession == "away")
	}

	return Summary{
		TotalTicks:         ticks,
		HomeGoals:          state.HomeGoals,
		AwayGoals:          state.AwayGoals,
		HomeAverageFatigue: averageFatigue(state.HomeFatigue),
		AwayAverageFatigue: averageFatigue(state.AwayFatigue),
		Events:             state.Events,
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

func resolveTickActions(seed int64, tick int, attacking, defending TeamPlan, possession string, state *State) []domain.MatchEvent {
	if !shouldCreateChance(seed, tick, attacking, defending) {
		return nil
	}

	minute := tickToMinute(tick)
	events := []domain.MatchEvent{
		makeEvent(tick, minute, "build_up", "middle-third", attacking, state),
	}

	phase := resolveAttackPhase(seed, tick, attacking, defending)
	if !phase.advanced {
		events = append(events, makeEvent(tick, minute, "turnover", phase.zone, attacking, state))
		return events
	}

	events = append(events, makeEvent(tick, minute, "penetration", phase.zone, attacking, state))
	switch resolveShotOutcome(seed, tick, attacking, defending) {
	case "goal":
		if possession == "home" {
			state.HomeGoals++
		} else {
			state.AwayGoals++
		}
		events = append(events, makeEvent(tick, minute, "shot", phase.zone, attacking, state))
		events = append(events, makeEvent(tick, minute, "goal", phase.zone, attacking, state))
	case "save":
		events = append(events, makeEvent(tick, minute, "shot", phase.zone, attacking, state))
		events = append(events, makeEvent(tick, minute, "save", phase.zone, defending, state))
	case "block":
		events = append(events, makeEvent(tick, minute, "shot", phase.zone, attacking, state))
		events = append(events, makeEvent(tick, minute, "block", phase.zone, defending, state))
	default:
		events = append(events, makeEvent(tick, minute, "turnover", phase.zone, attacking, state))
	}

	return events
}

func resolveAttackPhase(seed int64, tick int, attacking, defending TeamPlan) attackPhase {
	pressure := attacking.Control + attacking.Attack - defending.Defence
	value := (tick + int(seed%19) + max(pressure, -25)) % 9
	if value <= 3 {
		return attackPhase{zone: "half-space", advanced: true}
	}
	if value <= 5 {
		return attackPhase{zone: "final-third", advanced: true}
	}
	return attackPhase{zone: "middle-third", advanced: false}
}

func resolveShotOutcome(seed int64, tick int, attacking, defending TeamPlan) string {
	if shouldScore(seed, tick, attacking, defending) {
		return "goal"
	}
	value := (tick*2 + int(seed%23) + attacking.Attack - defending.Defence) % 6
	switch {
	case value <= 1:
		return "save"
	case value == 2:
		return "block"
	default:
		return "turnover"
	}
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

func makeEvent(tick, minute int, eventType, zone string, team TeamPlan, state *State) domain.MatchEvent {
	return domain.MatchEvent{
		Tick:      tick,
		Minute:    minute,
		Type:      eventType,
		PitchZone: zone,
		PayloadRaw: fmt.Sprintf(
			`{"team":"%s","home_goals":%d,"away_goals":%d}`,
			team.Club.ShortName,
			state.HomeGoals,
			state.AwayGoals,
		),
	}
}

func initialFatigue(lineup []Assignment) map[int64]float64 {
	fatigue := make(map[int64]float64, len(lineup))
	for _, assignment := range lineup {
		fatigue[assignment.Player.ID] = 0
	}
	return fatigue
}

func applyFatigueTick(fatigue map[int64]float64, lineup []Assignment, inPossession bool) {
	for _, assignment := range lineup {
		load := fatigueLoad(assignment)
		if inPossession {
			load += 0.015
		}
		fatigue[assignment.Player.ID] += load
	}
}

func fatigueLoad(assignment Assignment) float64 {
	if assignment.Player.PrimaryPosition == domain.PositionGK {
		return 0.008
	}
	switch assignment.Slot.Code {
	case "RB", "LB", "RW", "LW", "RM", "LM":
		return 0.032
	case "ST", "RST", "LST":
		return 0.028
	default:
		return 0.024
	}
}

func fatigueAdjustedPlan(plan TeamPlan, fatigue map[int64]float64) TeamPlan {
	adjusted := plan
	average := averageFatigue(fatigue)
	adjusted.Attack = max(20, plan.Attack-int(average/3.5))
	adjusted.Control = max(20, plan.Control-int(average/3.0))
	adjusted.Defence = max(20, plan.Defence-int(average/4.0))
	return adjusted
}

func averageFatigue(fatigue map[int64]float64) float64 {
	if len(fatigue) == 0 {
		return 0
	}
	total := 0.0
	for _, value := range fatigue {
		total += value
	}
	return total / float64(len(fatigue))
}
