package domain

import (
	"encoding/json"
	"testing"
)

func TestPlayerMarshalJSON(t *testing.T) {
	p := Player{
		ID:       "p1",
		Name:     "Test Player",
		Position: "FW",
		Club:     "Test FC",
		League:   "EPL",
		Rating:   7.5,
	}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal Player: %v", err)
	}
	var decoded Player
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal Player: %v", err)
	}
	if decoded.ID != p.ID || decoded.Name != p.Name || decoded.Position != p.Position || decoded.Rating != p.Rating {
		t.Fatalf("round-trip mismatch: %#v vs %#v", p, decoded)
	}
}

func TestLineupRoundTrip(t *testing.T) {
	l := Lineup{
		MatchID:   "m1",
		TeamID:    "t1",
		Formation: "4-3-3",
		Starters: []LineupSlot{
			{PlayerID: "p1", Position: "GK", ShirtNo: 1, IsCaptain: false},
			{PlayerID: "p2", Position: "FW", ShirtNo: 9, IsCaptain: true},
		},
		Bench: []LineupSlot{
			{PlayerID: "p3", Position: "MF", ShirtNo: 8, IsCaptain: false},
		},
	}
	data, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("marshal Lineup: %v", err)
	}
	var decoded Lineup
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal Lineup: %v", err)
	}
	if len(decoded.Starters) != 2 || len(decoded.Bench) != 1 {
		t.Fatalf("lineup slots mismatch: %#v", decoded)
	}
	if decoded.Starters[1].IsCaptain != true {
		t.Fatalf("captain flag lost")
	}
}

func TestMatchEventWithGoal(t *testing.T) {
	evt := MatchEvent{
		ID:       "e1",
		MatchID:  "m1",
		Minute:   34,
		Stoppage: 0,
		Kind:     EventGoal,
		Goal: &Goal{
			ScorerID:   "p1",
			ScorerName: "Test Player",
			GoalType:   "open_play",
			IsOwnGoal:  false,
			IsPenalty:  false,
		},
	}
	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("marshal MatchEvent: %v", err)
	}
	var decoded MatchEvent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal MatchEvent: %v", err)
	}
	if decoded.Kind != EventGoal || decoded.Goal == nil || decoded.Goal.ScorerID != "p1" {
		t.Fatalf("event round-trip mismatch: %#v", decoded)
	}
}

func TestMatchEventWithoutOptionalFields(t *testing.T) {
	evt := MatchEvent{
		ID:      "e2",
		MatchID: "m2",
		Minute:  67,
		Kind:    EventSubstitution,
		Sub: &Substitution{
			PlayerIn:  "p4",
			PlayerOut: "p2",
			Reason:    "tactical",
		},
	}
	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(data) == "" {
		t.Fatalf("empty JSON output")
	}
}

func TestTeamFormDefaults(t *testing.T) {
	tf := TeamForm{
		TeamID:        "t1",
		WindowMatches: 5,
	}
	if tf.FormRating != 0 {
		t.Fatalf("form rating should default to 0")
	}
	if tf.WindowMatches != 5 {
		t.Fatalf("window matches not set")
	}
}

func TestPlayerFormDefaults(t *testing.T) {
	pf := PlayerForm{
		PlayerID:      "p1",
		WindowMatches: 5,
	}
	if pf.FormRating != 0 {
		t.Fatalf("form rating should default to 0")
	}
}

func TestLoadPlayerStatsFromGlob(t *testing.T) {
	stats, used, err := LoadPlayerStatsFromGlob("testdata/player_stats_*.csv")
	if err != nil {
		t.Fatalf("LoadPlayerStatsFromGlob error: %v", err)
	}
	if len(used) != 1 {
		t.Fatalf("used files = %d, want 1", len(used))
	}
	if len(stats) != 6 {
		t.Fatalf("stats count = %d, want 6", len(stats))
	}
	if stats[0].PlayerID != "p1" {
		t.Fatalf("first player = %s, want p1", stats[0].PlayerID)
	}
	if stats[0].Goals != 1 || stats[0].Minutes != 90 {
		t.Fatalf("first player stats wrong: goals=%d minutes=%d", stats[0].Goals, stats[0].Minutes)
	}
}

func TestLoadPlayerStatsNoMatch(t *testing.T) {
	_, _, err := LoadPlayerStatsFromGlob("testdata/nonexistent_*.csv")
	if err == nil {
		t.Fatalf("expected error for no match")
	}
}
