package match

import (
	"fmt"
	"slices"

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
	Seed         int64
	Tick         int
	HomeGoals    int
	AwayGoals    int
	Possession   string
	HomeFatigue  map[int64]float64
	AwayFatigue  map[int64]float64
	HomeInjured  map[int64]Injury
	AwayInjured  map[int64]Injury
	HomeBookings map[int64]int
	AwayBookings map[int64]int
	HomeRedCards map[int64]CardRecord
	AwayRedCards map[int64]CardRecord
	Cards        []CardRecord
	Suspensions  []Suspension
	Events       []domain.MatchEvent
}

// Summary is the terminal output of the tick loop.
type Summary struct {
	TotalTicks         int
	HomeGoals          int
	AwayGoals          int
	HomeAverageFatigue float64
	AwayAverageFatigue float64
	Injuries           []Injury
	Cards              []CardRecord
	Suspensions        []Suspension
	Events             []domain.MatchEvent
}

type attackPhase struct {
	zone     string
	advanced bool
}

// Injury is a deterministic in-match injury record.
type Injury struct {
	PlayerID int64
	Team     string
	Tick     int
	Severity string
}

// CardRecord captures an in-match booking or dismissal.
type CardRecord struct {
	PlayerID int64
	Team     string
	Tick     int
	Card     string
}

// Suspension represents an immediate next-match ban created in the match.
type Suspension struct {
	PlayerID int64
	Team     string
	Reason   string
	Matches  int
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
		Seed:         seed,
		Possession:   "home",
		HomeFatigue:  initialFatigue(home.Lineup),
		AwayFatigue:  initialFatigue(away.Lineup),
		HomeInjured:  make(map[int64]Injury),
		AwayInjured:  make(map[int64]Injury),
		HomeBookings: make(map[int64]int),
		AwayBookings: make(map[int64]int),
		HomeRedCards: make(map[int64]CardRecord),
		AwayRedCards: make(map[int64]CardRecord),
		Cards:        make([]CardRecord, 0, ticks/10),
		Suspensions:  make([]Suspension, 0, 4),
		Events:       make([]domain.MatchEvent, 0, ticks/6),
	}

	for tick := 1; tick <= ticks; tick++ {
		state.Tick = tick
		effectiveHome := fatigueAdjustedPlan(home, state.HomeFatigue, len(state.HomeInjured)+len(state.HomeRedCards))
		effectiveAway := fatigueAdjustedPlan(away, state.AwayFatigue, len(state.AwayInjured)+len(state.AwayRedCards))
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
		for _, event := range resolveInjuryEvents(seed, tick, home, away, &state) {
			state.Events = append(state.Events, event)
		}
		for _, event := range resolveCardEvents(seed, tick, home, away, &state) {
			state.Events = append(state.Events, event)
		}
	}

	return Summary{
		TotalTicks:         ticks,
		HomeGoals:          state.HomeGoals,
		AwayGoals:          state.AwayGoals,
		HomeAverageFatigue: averageFatigue(home.Lineup, state.HomeFatigue),
		AwayAverageFatigue: averageFatigue(away.Lineup, state.AwayFatigue),
		Injuries:           flattenInjuries(state.HomeInjured, state.AwayInjured),
		Cards:              append([]CardRecord(nil), state.Cards...),
		Suspensions:        append([]Suspension(nil), state.Suspensions...),
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

func fatigueAdjustedPlan(plan TeamPlan, fatigue map[int64]float64, injuredCount int) TeamPlan {
	adjusted := plan
	average := averageFatigue(plan.Lineup, fatigue)
	injuryPenalty := injuredCount * 6
	adjusted.Attack = max(20, plan.Attack-int(average/3.5)-injuryPenalty)
	adjusted.Control = max(20, plan.Control-int(average/3.0)-injuryPenalty)
	adjusted.Defence = max(20, plan.Defence-int(average/4.0)-injuryPenalty)
	return adjusted
}

func averageFatigue(lineup []Assignment, fatigue map[int64]float64) float64 {
	if len(lineup) == 0 {
		return 0
	}
	total := 0.0
	for _, assignment := range lineup {
		total += fatigue[assignment.Player.ID]
	}
	return total / float64(len(lineup))
}

func resolveInjuryEvents(seed int64, tick int, home, away TeamPlan, state *State) []domain.MatchEvent {
	events := make([]domain.MatchEvent, 0, 2)
	if injury, ok := maybeInjureTeam(seed, tick, home, state.HomeFatigue, state.HomeInjured); ok {
		state.HomeInjured[injury.PlayerID] = injury
		events = append(events, injuryEvent(injury, home, state))
	}
	if injury, ok := maybeInjureTeam(seed+11, tick, away, state.AwayFatigue, state.AwayInjured); ok {
		state.AwayInjured[injury.PlayerID] = injury
		events = append(events, injuryEvent(injury, away, state))
	}
	return events
}

func maybeInjureTeam(seed int64, tick int, team TeamPlan, fatigue map[int64]float64, injured map[int64]Injury) (Injury, bool) {
	assignment, load, ok := highestFatigueCandidate(team.Lineup, fatigue, injured)
	if !ok {
		return Injury{}, false
	}
	if load < 2.4 {
		return Injury{}, false
	}
	trigger := (tick + int(seed%31) + int(load*10)) % 41
	if trigger != 0 {
		return Injury{}, false
	}
	return Injury{
		PlayerID: assignment.Player.ID,
		Team:     team.Club.ShortName,
		Tick:     tick,
		Severity: injurySeverity(load),
	}, true
}

func highestFatigueCandidate(lineup []Assignment, fatigue map[int64]float64, injured map[int64]Injury) (Assignment, float64, bool) {
	bestLoad := -1.0
	var best Assignment
	found := false
	for _, assignment := range lineup {
		if assignment.Player.PrimaryPosition == domain.PositionGK {
			continue
		}
		if _, already := injured[assignment.Player.ID]; already {
			continue
		}
		load := fatigue[assignment.Player.ID]
		if !found || load > bestLoad || (load == bestLoad && comparePlayers(assignment.Player, best.Player) < 0) {
			best = assignment
			bestLoad = load
			found = true
		}
	}
	return best, bestLoad, found
}

func injurySeverity(load float64) string {
	switch {
	case load >= 5.0:
		return "major"
	case load >= 3.5:
		return "moderate"
	default:
		return "minor"
	}
}

func injuryEvent(injury Injury, team TeamPlan, state *State) domain.MatchEvent {
	return domain.MatchEvent{
		Tick:      injury.Tick,
		Minute:    tickToMinute(injury.Tick),
		Type:      "injury",
		PitchZone: "recovery-phase",
		PayloadRaw: fmt.Sprintf(
			`{"team":"%s","player_id":%d,"severity":"%s","home_goals":%d,"away_goals":%d}`,
			team.Club.ShortName,
			injury.PlayerID,
			injury.Severity,
			state.HomeGoals,
			state.AwayGoals,
		),
	}
}

func flattenInjuries(home, away map[int64]Injury) []Injury {
	total := make([]Injury, 0, len(home)+len(away))
	for _, injury := range home {
		total = append(total, injury)
	}
	for _, injury := range away {
		total = append(total, injury)
	}
	slices.SortFunc(total, func(left, right Injury) int {
		if left.Tick != right.Tick {
			if left.Tick < right.Tick {
				return -1
			}
			return 1
		}
		if left.Team != right.Team {
			if left.Team < right.Team {
				return -1
			}
			return 1
		}
		if left.PlayerID < right.PlayerID {
			return -1
		}
		if left.PlayerID > right.PlayerID {
			return 1
		}
		return 0
	})
	return total
}

func resolveCardEvents(seed int64, tick int, home, away TeamPlan, state *State) []domain.MatchEvent {
	events := make([]domain.MatchEvent, 0, 2)
	if record, ok := maybeCardTeam(seed, tick, home, state.HomeFatigue, state.HomeBookings, state.HomeRedCards, state.HomeInjured); ok {
		applyCardRecord(record, state.HomeBookings, state.HomeRedCards, &state.Cards, &state.Suspensions)
		events = append(events, cardEvent(record, state))
	}
	if record, ok := maybeCardTeam(seed+7, tick, away, state.AwayFatigue, state.AwayBookings, state.AwayRedCards, state.AwayInjured); ok {
		applyCardRecord(record, state.AwayBookings, state.AwayRedCards, &state.Cards, &state.Suspensions)
		events = append(events, cardEvent(record, state))
	}
	return events
}

func maybeCardTeam(seed int64, tick int, team TeamPlan, fatigue map[int64]float64, bookings map[int64]int, redCards map[int64]CardRecord, injured map[int64]Injury) (CardRecord, bool) {
	assignment, load, ok := highestFatigueCandidate(team.Lineup, fatigue, injured)
	if !ok {
		return CardRecord{}, false
	}
	if _, sentOff := redCards[assignment.Player.ID]; sentOff {
		return CardRecord{}, false
	}
	trigger := (tick + int(seed%17) + int(load*10) + assignment.Player.Attributes.Defending) % 29
	if trigger != 0 {
		return CardRecord{}, false
	}

	card := "yellow"
	if bookings[assignment.Player.ID] >= 1 || load >= 4.8 {
		card = "red"
	}
	return CardRecord{
		PlayerID: assignment.Player.ID,
		Team:     team.Club.ShortName,
		Tick:     tick,
		Card:     card,
	}, true
}

func applyCardRecord(record CardRecord, bookings map[int64]int, redCards map[int64]CardRecord, cards *[]CardRecord, suspensions *[]Suspension) {
	if record.Card == "yellow" {
		bookings[record.PlayerID]++
		*cards = append(*cards, record)
		return
	}
	bookings[record.PlayerID] = 2
	redCards[record.PlayerID] = record
	*cards = append(*cards, record)
	*suspensions = append(*suspensions, Suspension{
		PlayerID: record.PlayerID,
		Team:     record.Team,
		Reason:   "red_card",
		Matches:  1,
	})
}

func cardEvent(record CardRecord, state *State) domain.MatchEvent {
	return domain.MatchEvent{
		Tick:      record.Tick,
		Minute:    tickToMinute(record.Tick),
		Type:      record.Card,
		PitchZone: "defensive-third",
		PayloadRaw: fmt.Sprintf(
			`{"team":"%s","player_id":%d,"card":"%s","home_goals":%d,"away_goals":%d}`,
			record.Team,
			record.PlayerID,
			record.Card,
			state.HomeGoals,
			state.AwayGoals,
		),
	}
}
