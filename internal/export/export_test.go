package export

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ismael/football-analytics/internal/domain"
	"github.com/ismael/football-analytics/internal/engine/season"
)

func TestExportPlayersCSV_header(t *testing.T) {
	cid := int64(1)
	players := []domain.Player{
		{
			ID: 1, ClubID: &cid,
			FirstName: "Test", LastName: "Player",
			PrimaryPosition: domain.PositionST,
			Attributes: domain.PlayerAttributes{Overall: 82, Potential: 85, Pace: 80, Shooting: 78, Passing: 72, Defending: 45, Keeping: 40},
		},
	}
	var buf bytes.Buffer
	if err := ExportPlayersCSV(players, &buf); err != nil {
		t.Fatalf("export: %v", err)
	}
	output := buf.String()
	for _, field := range []string{"id", "name", "position", "overall"} {
		if !strings.Contains(output, field) {
			t.Fatalf("CSV missing column %q", field)
		}
	}
	if !strings.Contains(output, "Test Player") {
		t.Fatal("CSV missing player name")
	}
}

func TestExportStandingsCSV_header(t *testing.T) {
	table := []season.Standing{
		{ClubID: 1, ClubName: "Arsenal FC", Played: 5, Won: 4, Drawn: 1, Lost: 0, GoalsFor: 12, GoalsAgainst: 3, GoalDifference: 9, Points: 13},
	}
	var buf bytes.Buffer
	if err := ExportStandingsCSV(table, &buf); err != nil {
		t.Fatalf("export: %v", err)
	}
	output := buf.String()
	for _, field := range []string{"position", "club", "points"} {
		if !strings.Contains(output, field) {
			t.Fatalf("CSV missing column %q", field)
		}
	}
	if !strings.Contains(output, "Arsenal FC") {
		t.Fatal("CSV missing club name")
	}
}

func TestExportMatchResultsJSON_structure(t *testing.T) {
	results := []season.MatchResult{
		{
			Fixture: domain.Fixture{Matchday: 1, HomeClubID: 1, AwayClubID: 2},
		},
	}
	var buf bytes.Buffer
	if err := ExportMatchResultsJSON(results, &buf); err != nil {
		t.Fatalf("export: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "matchday") || !strings.Contains(output, "home_goals") {
		t.Fatalf("JSON missing expected fields: %s", output)
	}
}
