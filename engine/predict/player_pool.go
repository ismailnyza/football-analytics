package predict

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sort"
)

type PoolPlayer struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Team     string  `json:"team"`
	Position string  `json:"position"`
	Rating   float64 `json:"rating"`
}

type PlayerPool struct {
	Players   []PoolPlayer            `json:"players"`
	ByTeam    map[string][]PoolPlayer `json:"by_team"`
	TeamTotal map[string]float64      `json:"team_total"`
}

func NewPlayerPool(teams map[string]float64) *PlayerPool {
	pool := &PlayerPool{
		ByTeam:    make(map[string][]PoolPlayer),
		TeamTotal: make(map[string]float64),
	}
	rng := rand.New(rand.NewPCG(42, 42))
	playerID := 0

	for team, rating := range teams {
		positions := []struct {
			pos   string
			count int
			bias  float64
		}{
			{"GK", 2, -150},
			{"DF", 6, -30},
			{"MF", 6, -10},
			{"FW", 6, 0},
		}
		var teamPlayers []PoolPlayer
		var teamTotal float64
		for _, p := range positions {
			for i := 0; i < p.count; i++ {
				playerID++
				noise := (rng.Float64() - 0.5) * 160
				playerRating := rating + p.bias + noise
				playerRating = math.Max(playerRating, 1200)
				playerRating = math.Min(playerRating, 2000)
				pl := PoolPlayer{
					ID:       fmt.Sprintf("P%03d", playerID),
					Name:     fmt.Sprintf("%s-%s-%d", teamAbbrev(team), p.pos, i+1),
					Team:     team,
					Position: p.pos,
					Rating:   playerRating,
				}
				teamPlayers = append(teamPlayers, pl)
				teamTotal += playerRating
			}
		}
		sort.Slice(teamPlayers, func(i, j int) bool {
			return teamPlayers[i].Rating > teamPlayers[j].Rating
		})
		pool.ByTeam[team] = teamPlayers
		pool.TeamTotal[team] = teamTotal
		pool.Players = append(pool.Players, teamPlayers...)
	}
	return pool
}

func (p *PlayerPool) GetTeamPlayers(team string) []PoolPlayer {
	return p.ByTeam[team]
}

func (p *PlayerPool) ScorerProbabilities(team string, goals int) []ScorerPrediction {
	players := p.ByTeam[team]
	total := p.TeamTotal[team]
	if total == 0 || len(players) == 0 {
		return nil
	}

	fwPlayers := filterByPosition(players, "FW")
	mfPlayers := filterByPosition(players, "MF")
	dfPlayers := filterByPosition(players, "DF")
	weightTotal := ratingSum(fwPlayers)*2.0 + ratingSum(mfPlayers)*0.8 + ratingSum(dfPlayers)*0.2

	var preds []ScorerPrediction
	for _, pl := range players {
		var weight float64
		switch pl.Position {
		case "FW":
			weight = pl.Rating * 2.0
		case "MF":
			weight = pl.Rating * 0.8
		case "DF":
			weight = pl.Rating * 0.2
		default:
			weight = pl.Rating * 0.01
		}
		prob := weight / weightTotal
		if prob > 0.001 {
			preds = append(preds, ScorerPrediction{
				Player:          pl,
				GoalProbability: prob * float64(goals),
				ExpectedGoals:   prob * float64(goals),
			})
		}
	}
	_ = total
	sort.Slice(preds, func(i, j int) bool {
		return preds[i].GoalProbability > preds[j].GoalProbability
	})
	return preds
}

func (p *PlayerPool) AssistProbabilities(team string, goals int) []AssistPrediction {
	players := p.ByTeam[team]
	var weighted []struct {
		player PoolPlayer
		weight float64
	}
	totalWeight := 0.0
	for _, pl := range players {
		w := assistWeight(pl.Position, pl.Rating)
		if w > 0 {
			weighted = append(weighted, struct {
				player PoolPlayer
				weight float64
			}{pl, w})
			totalWeight += w
		}
	}
	expectedAssists := float64(goals) * 0.7
	var preds []AssistPrediction
	for _, w := range weighted {
		if totalWeight > 0 {
			preds = append(preds, AssistPrediction{
				Player:            w.player,
				AssistProbability: w.weight / totalWeight * expectedAssists,
				ExpectedAssists:   w.weight / totalWeight * expectedAssists,
			})
		}
	}
	sort.Slice(preds, func(i, j int) bool {
		return preds[i].AssistProbability > preds[j].AssistProbability
	})
	return preds
}

func assistWeight(pos string, rating float64) float64 {
	switch pos {
	case "MF":
		return rating * 1.5
	case "FW":
		return rating * 0.8
	case "DF":
		return rating * 0.3
	default:
		return 0
	}
}

func filterByPosition(players []PoolPlayer, pos string) []PoolPlayer {
	var result []PoolPlayer
	for _, p := range players {
		if p.Position == pos {
			result = append(result, p)
		}
	}
	return result
}

func ratingSum(players []PoolPlayer) float64 {
	var s float64
	for _, p := range players {
		s += p.Rating
	}
	return s
}

func teamAbbrev(team string) string {
	if len(team) <= 3 {
		return team
	}
	return team[:3]
}
