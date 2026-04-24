package predict

import (
	"fmt"
	"math/rand/v2"
	"sort"

	"github.com/ismailnyza/football-analytics/engine/domain"
)

type RichTeamStats struct {
	Team          string
	Shots         int
	ShotsOnTarget int
	Fouls         int
	Corners       int
	YellowCards   int
	RedCards      int
	HalfTimeGoals int
	FullTimeGoals int
}

type ReconstructedMatch struct {
	HomeTeam  string              `json:"home_team"`
	AwayTeam  string              `json:"away_team"`
	Scoreline string              `json:"scoreline"`
	HalfTime  string              `json:"half_time"`
	Events    []domain.MatchEvent `json:"events"`
	Stats     struct {
		Home RichTeamStats `json:"home"`
		Away RichTeamStats `json:"away"`
	} `json:"stats"`
}

func ReconstructMatch(home, away string, homeGoals, awayGoals int, homePool, awayPool []PoolPlayer, homeStats, awayStats RichTeamStats) ReconstructedMatch {
	rng := rand.New(rand.NewPCG(99, 99))
	matchID := fmt.Sprintf("%s_vs_%s", sanitize(home), sanitize(away))
	eventID := 0
	nextID := func() string { eventID++; return fmt.Sprintf("E%04d", eventID) }

	var events []domain.MatchEvent
	homeScored := 0
	awayScored := 0

	homeScorers := topScoringPlayers(homePool)
	awayScorers := topScoringPlayers(awayPool)

	homeAssisters := topAssistingPlayers(homePool)
	awayAssisters := topAssistingPlayers(awayPool)

	homeScorerIdx := 0
	awayScorerIdx := 0
	homeAssistIdx := 0
	awayAssistIdx := 0

	allGoalMinutes := generateGoalMinutes(homeGoals+awayGoals, rng)

	for _, minute := range allGoalMinutes {
		isHome := false
		if homeScored < homeGoals && awayScored < awayGoals {
			isHome = rng.Float64() < float64(homeGoals)/float64(homeGoals+awayGoals)
		} else if homeScored < homeGoals {
			isHome = true
		} else {
			isHome = false
		}

		if isHome {
			homeScored++
			scorer := homeScorers[homeScorerIdx%len(homeScorers)]
			homeScorerIdx++
			assister := homeAssisters[homeAssistIdx%len(homeAssisters)]
			homeAssistIdx++

			events = append(events, domain.MatchEvent{
				ID:      nextID(),
				MatchID: matchID,
				Minute:  minute,
				Kind:    domain.EventGoal,
				Goal: &domain.Goal{
					ScorerID:   scorer.ID,
					ScorerName: scorer.Name,
					GoalType:   randomGoalType(rng),
					IsPenalty:  rng.Float64() < 0.1,
				},
			})
			events = append(events, domain.MatchEvent{
				ID:      nextID(),
				MatchID: matchID,
				Minute:  minute,
				Kind:    domain.EventAssist,
			})
			_ = assister
		} else {
			awayScored++
			scorer := awayScorers[awayScorerIdx%len(awayScorers)]
			awayScorerIdx++
			assister := awayAssisters[awayAssistIdx%len(awayAssisters)]
			awayAssistIdx++

			events = append(events, domain.MatchEvent{
				ID:      nextID(),
				MatchID: matchID,
				Minute:  minute,
				Kind:    domain.EventGoal,
				Goal: &domain.Goal{
					ScorerID:   scorer.ID,
					ScorerName: scorer.Name,
					GoalType:   randomGoalType(rng),
					IsPenalty:  rng.Float64() < 0.1,
				},
			})
			events = append(events, domain.MatchEvent{
				ID:      nextID(),
				MatchID: matchID,
				Minute:  minute,
				Kind:    domain.EventAssist,
			})
			_ = assister
		}
	}

	cardEvents := generateCardEvents(matchID, home, away, homeStats.YellowCards, awayStats.YellowCards, awayPool, rng, nextID, len(events))
	events = append(events, cardEvents...)

	sort.Slice(events, func(i, j int) bool {
		return events[i].Minute < events[j].Minute
	})

	htHome := countGoalsBefore(events, home, 45)
	htAway := countGoalsBefore(events, away, 45)

	return ReconstructedMatch{
		HomeTeam:  home,
		AwayTeam:  away,
		Scoreline: fmt.Sprintf("%d-%d", homeGoals, awayGoals),
		HalfTime:  fmt.Sprintf("%d-%d", htHome, htAway),
		Events:    events,
		Stats: struct {
			Home RichTeamStats `json:"home"`
			Away RichTeamStats `json:"away"`
		}{Home: homeStats, Away: awayStats},
	}
}

func topScoringPlayers(pool []PoolPlayer) []PoolPlayer {
	if len(pool) == 0 {
		return []PoolPlayer{{ID: "P000", Name: "Unknown", Team: "Unknown", Position: "FW", Rating: 1500}}
	}
	var fws []PoolPlayer
	for _, p := range pool {
		if p.Position == "FW" {
			fws = append(fws, p)
		}
	}
	sort.Slice(fws, func(i, j int) bool { return fws[i].Rating > fws[j].Rating })
	if len(fws) == 0 {
		return pool
	}
	return fws
}

func topAssistingPlayers(pool []PoolPlayer) []PoolPlayer {
	if len(pool) == 0 {
		return []PoolPlayer{{ID: "P000", Name: "Unknown", Team: "Unknown", Position: "MF", Rating: 1500}}
	}
	var mfs []PoolPlayer
	for _, p := range pool {
		if p.Position == "MF" {
			mfs = append(mfs, p)
		}
	}
	sort.Slice(mfs, func(i, j int) bool { return mfs[i].Rating > mfs[j].Rating })
	if len(mfs) == 0 {
		return pool
	}
	return mfs
}

func generateGoalMinutes(totalGoals int, rng *rand.Rand) []int {
	minutes := make([]int, totalGoals)
	for i := range minutes {
		minutes[i] = 1 + int(rng.Float64()*90)
	}
	sort.Ints(minutes)
	return minutes
}

func generateCardEvents(matchID, home, away string, homeYellows, awayYellows int, awayPool []PoolPlayer, rng *rand.Rand, nextID func() string, baseID int) []domain.MatchEvent {
	var events []domain.MatchEvent
	for i := 0; i < homeYellows; i++ {
		minute := 1 + int(rng.Float64()*90)
		events = append(events, domain.MatchEvent{
			ID:      nextID(),
			MatchID: matchID,
			Minute:  minute,
			Kind:    domain.EventCard,
			Card:    &domain.Card{PlayerID: fmt.Sprintf("%s-DF-%d", sanitize(home), i+1), CardType: "yellow", Reason: "foul"},
		})
	}
	for i := 0; i < awayYellows; i++ {
		if i < len(awayPool) {
			minute := 1 + int(rng.Float64()*90)
			events = append(events, domain.MatchEvent{
				ID:      nextID(),
				MatchID: matchID,
				Minute:  minute,
				Kind:    domain.EventCard,
				Card:    &domain.Card{PlayerID: awayPool[i].ID, CardType: "yellow", Reason: "foul"},
			})
		}
	}
	_ = baseID
	return events
}

func countGoalsBefore(events []domain.MatchEvent, team string, beforeMinute int) int {
	count := 0
	for _, e := range events {
		if e.Kind == domain.EventGoal && e.Minute <= beforeMinute {
			count++
		}
	}
	return count / 2
}

func randomGoalType(rng *rand.Rand) string {
	types := []string{"open_play", "header", "penalty", "free_kick", "long_range"}
	return types[rng.IntN(len(types))]
}

func sanitize(s string) string {
	result := ""
	for _, c := range s {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			result += string(c)
		}
	}
	if result == "" {
		return "TEAM"
	}
	return result
}

func PredictMatchWithEvents(home, away string, homeGoals, awayGoals int, pool *PlayerPool, homeStats, awayStats RichTeamStats) ReconstructedMatch {
	homePlayers := pool.GetTeamPlayers(home)
	awayPlayers := pool.GetTeamPlayers(away)
	return ReconstructMatch(home, away, homeGoals, awayGoals, homePlayers, awayPlayers, homeStats, awayStats)
}
