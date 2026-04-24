package domain

type Player struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Position string  `json:"position"`
	Club     string  `json:"club"`
	League   string  `json:"league"`
	Rating   float64 `json:"rating"`
}

type PlayerRating struct {
	PlayerID string  `json:"player_id"`
	Rating   float64 `json:"rating"`
	Date     string  `json:"date"`
	Source   string  `json:"source"`
}

type PlayerForm struct {
	PlayerID      string  `json:"player_id"`
	RecentGoals   int     `json:"recent_goals"`
	RecentAssists int     `json:"recent_assists"`
	RecentMatches int     `json:"recent_matches"`
	FormRating    float64 `json:"form_rating"`
	WindowMatches int     `json:"window_matches"`
}

type PlayerMatchStats struct {
	PlayerID       string  `json:"player_id"`
	MatchID        string  `json:"match_id"`
	Season         string  `json:"season"`
	Date           string  `json:"date"`
	Minutes        int     `json:"minutes"`
	Goals          int     `json:"goals"`
	Assists        int     `json:"assists"`
	Shots          int     `json:"shots"`
	ShotsOnTarget  int     `json:"shots_on_target"`
	XG             float64 `json:"xg"`
	XA             float64 `json:"xa"`
	Passes         int     `json:"passes"`
	KeyPasses      int     `json:"key_passes"`
	Tackles        int     `json:"tackles"`
	Interceptions  int     `json:"interceptions"`
	Fouls          int     `json:"fouls"`
	Fouled         int     `json:"fouled"`
	YellowCards    int     `json:"yellow_cards"`
	RedCards       int     `json:"red_cards"`
	PositionPlayed string  `json:"position_played"`
	SourceFile     string  `json:"source_file"`
}
