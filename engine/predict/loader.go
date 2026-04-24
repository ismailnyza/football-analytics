package predict

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type PlayerData struct {
	Name   string  `json:"name"`
	Pos    string  `json:"pos"`
	Rating float64 `json:"rating"`
}

func LoadPlayerPoolFromJSON(path string) (*PlayerPool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read player data: %w", err)
	}
	raw := make(map[string][]PlayerData)
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal player data: %w", err)
	}

	pool := &PlayerPool{
		ByTeam:    make(map[string][]PoolPlayer),
		TeamTotal: make(map[string]float64),
	}
	playerID := 0

	for team, players := range raw {
		var teamPlayers []PoolPlayer
		var teamTotal float64
		for _, pd := range players {
			playerID++
			pl := PoolPlayer{
				ID:       fmt.Sprintf("P%04d", playerID),
				Name:     pd.Name,
				Team:     team,
				Position: pd.Pos,
				Rating:   pd.Rating,
			}
			teamPlayers = append(teamPlayers, pl)
			teamTotal += pd.Rating
		}
		sort.Slice(teamPlayers, func(i, j int) bool {
			return teamPlayers[i].Rating > teamPlayers[j].Rating
		})
		pool.ByTeam[team] = teamPlayers
		pool.TeamTotal[team] = teamTotal
		pool.Players = append(pool.Players, teamPlayers...)
	}
	return pool, nil
}
