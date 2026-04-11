package screens

import (
	"fmt"

	"github.com/ismael/football-analytics/internal/tui/screens/matchlab"
)

var copyBySection = map[string]string{
	"Dashboard":     "Club overview, world summary, and season entry points will live here.",
	"Season Lab":    "Fixture simulation, standings progression, and branchable season runs.",
	"Club":          "Club profile, finances, tactics, and strategic decisions.",
	"Squad":         "Roster management, role balance, and availability snapshots.",
	"Player Detail": "Player identity, attributes, form, and longitudinal development.",
	"Transfers":     "Transfer shortlist, valuation, negotiations, and contract actions.",
	"World":         "Global competitions, branches, and long-term simulation state.",
	"Scenarios":     "Alternate timelines, saved checkpoints, and replay branches.",
	"Data / Import": "Ingestion status, source validation, and publish workflows.",
	"Settings":      "Runtime configuration, display options, and simulation defaults.",
}

// ContentFor returns placeholder copy for screens that are not yet interactive.
func ContentFor(section string) string {
	if copy, ok := copyBySection[section]; ok {
		return fmt.Sprintf("%s\n\nStatus: planned shell section", copy)
	}
	return "Unknown section"
}

// NewScreenFor returns a live interactive Screen for sections that have one,
// or nil for sections that still render as static placeholder text.
func NewScreenFor(section string) Screen {
	switch section {
	case "Match Lab":
		m := matchlab.New()
		return &matchlabAdapter{model: m}
	}
	return nil
}
