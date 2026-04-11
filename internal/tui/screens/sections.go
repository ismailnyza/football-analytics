package screens

import "fmt"

var copyBySection = map[string]string{
	"Dashboard":     "Club overview, world summary, and season entry points will live here.",
	"Match Lab":     "Single-match scenarios, lineup experiments, and deterministic replay controls.",
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

// ContentFor returns the current placeholder copy for a primary screen.
func ContentFor(section string) string {
	if copy, ok := copyBySection[section]; ok {
		return fmt.Sprintf("%s\n\nStatus: planned shell section", copy)
	}
	return "Unknown section"
}
