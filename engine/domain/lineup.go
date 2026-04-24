package domain

type Lineup struct {
	MatchID   string       `json:"match_id"`
	TeamID    string       `json:"team_id"`
	Formation string       `json:"formation"`
	Starters  []LineupSlot `json:"starters"`
	Bench     []LineupSlot `json:"bench"`
}

type LineupSlot struct {
	PlayerID  string `json:"player_id"`
	Position  string `json:"position"`
	ShirtNo   int    `json:"shirt_no"`
	IsCaptain bool   `json:"is_captain"`
}
