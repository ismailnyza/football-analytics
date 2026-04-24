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
