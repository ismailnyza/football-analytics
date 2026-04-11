// Package export provides CSV and JSON export tools for simulation entities.
package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/ismael/football-analytics/internal/domain"
	"github.com/ismael/football-analytics/internal/engine/season"
)

// ExportPlayersCSV writes a CSV of player attributes to w.
func ExportPlayersCSV(players []domain.Player, w io.Writer) error {
	cw := csv.NewWriter(w)
	header := []string{"id", "name", "position", "overall", "potential", "pace", "shooting", "passing", "defending", "keeping"}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("export players csv header: %w", err)
	}
	for _, p := range players {
		row := []string{
			strconv.FormatInt(p.ID, 10),
			p.DisplayName(),
			string(p.PrimaryPosition),
			strconv.Itoa(p.Attributes.Overall),
			strconv.Itoa(p.Attributes.Potential),
			strconv.Itoa(p.Attributes.Pace),
			strconv.Itoa(p.Attributes.Shooting),
			strconv.Itoa(p.Attributes.Passing),
			strconv.Itoa(p.Attributes.Defending),
			strconv.Itoa(p.Attributes.Keeping),
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("export players csv row: %w", err)
		}
	}
	cw.Flush()
	return cw.Error()
}

// ExportStandingsCSV writes a league table to w as CSV.
func ExportStandingsCSV(table []season.Standing, w io.Writer) error {
	cw := csv.NewWriter(w)
	header := []string{"position", "club", "played", "won", "drawn", "lost", "gf", "ga", "gd", "points"}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("export standings header: %w", err)
	}
	for i, s := range table {
		row := []string{
			strconv.Itoa(i + 1),
			s.ClubName,
			strconv.Itoa(s.Played),
			strconv.Itoa(s.Won),
			strconv.Itoa(s.Drawn),
			strconv.Itoa(s.Lost),
			strconv.Itoa(s.GoalsFor),
			strconv.Itoa(s.GoalsAgainst),
			strconv.Itoa(s.GoalDifference),
			strconv.Itoa(s.Points),
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("export standings row: %w", err)
		}
	}
	cw.Flush()
	return cw.Error()
}

// ExportMatchResultsJSON marshals match results to JSON and writes to w.
func ExportMatchResultsJSON(results []season.MatchResult, w io.Writer) error {
	type row struct {
		Matchday  int    `json:"matchday"`
		HomeClub  int64  `json:"home_club_id"`
		AwayClub  int64  `json:"away_club_id"`
		HomeGoals int    `json:"home_goals"`
		AwayGoals int    `json:"away_goals"`
	}
	rows := make([]row, len(results))
	for i, r := range results {
		rows[i] = row{
			Matchday:  r.Fixture.Matchday,
			HomeClub:  r.Fixture.HomeClubID,
			AwayClub:  r.Fixture.AwayClubID,
			HomeGoals: r.Summary.HomeGoals,
			AwayGoals: r.Summary.AwayGoals,
		}
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(rows); err != nil {
		return fmt.Errorf("export match results json: %w", err)
	}
	return nil
}
