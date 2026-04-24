package domain

type Team struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	League   string  `json:"league"`
	Strength float64 `json:"strength"`
}

type TeamForm struct {
	TeamID             string  `json:"team_id"`
	RecentWins         int     `json:"recent_wins"`
	RecentDraws        int     `json:"recent_draws"`
	RecentLosses       int     `json:"recent_losses"`
	RecentGoalsFor     int     `json:"recent_goals_for"`
	RecentGoalsAgainst int     `json:"recent_goals_against"`
	FormRating         float64 `json:"form_rating"`
	WindowMatches      int     `json:"window_matches"`
}
