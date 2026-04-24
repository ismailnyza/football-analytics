package predict

type MatchPrediction struct {
	HomeTeam string `json:"home_team"`
	AwayTeam string `json:"away_team"`

	ResultProbabilities ResultProbabilities `json:"result_probabilities"`
	PredictedResult     string              `json:"predicted_result"`

	ExpectedHomeGoals  float64 `json:"expected_home_goals"`
	ExpectedAwayGoals  float64 `json:"expected_away_goals"`
	PredictedScoreline string  `json:"predicted_scoreline"`

	HomeScorers []ScorerPrediction `json:"home_scorers"`
	AwayScorers []ScorerPrediction `json:"away_scorers"`
	HomeAssists []AssistPrediction `json:"home_assists"`
	AwayAssists []AssistPrediction `json:"away_assists"`

	TeamEvents []TeamEventSummary `json:"team_events"`

	Reconstructed *ReconstructedMatch `json:"reconstructed,omitempty"`

	HomeForm *TeamFormSnapshot `json:"home_form,omitempty"`
	AwayForm *TeamFormSnapshot `json:"away_form,omitempty"`

	PlayerPoolSize int `json:"player_pool_size"`
}

type ResultProbabilities struct {
	HomeWin float64 `json:"home_win"`
	Draw    float64 `json:"draw"`
	AwayWin float64 `json:"away_win"`
}

type ScorerPrediction struct {
	Player          PoolPlayer `json:"player"`
	GoalProbability float64    `json:"goal_probability"`
	ExpectedGoals   float64    `json:"expected_goals"`
}

type AssistPrediction struct {
	Player            PoolPlayer `json:"player"`
	AssistProbability float64    `json:"assist_probability"`
	ExpectedAssists   float64    `json:"expected_assists"`
}

type TeamEventSummary struct {
	Team          string  `json:"team"`
	Shots         float64 `json:"expected_shots"`
	ShotsOnTarget float64 `json:"expected_shots_on_target"`
	Corners       float64 `json:"expected_corners"`
	Fouls         float64 `json:"expected_fouls"`
	YellowCards   float64 `json:"expected_yellow_cards"`
	RedCards      float64 `json:"expected_red_cards"`
}

type TeamFormSnapshot struct {
	Team          string  `json:"team"`
	RecentWins    int     `json:"recent_wins"`
	RecentDraws   int     `json:"recent_draws"`
	RecentLosses  int     `json:"recent_losses"`
	PointsPerGame float64 `json:"points_per_game"`
	GoalDiff      float64 `json:"goal_diff_per_game"`
	RecentGF      int     `json:"recent_goals_for"`
	RecentGA      int     `json:"recent_goals_against"`
}
