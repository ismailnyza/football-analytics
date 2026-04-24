package baseline

import "time"

type Match struct {
	Competition string    `json:"competition"`
	Season      string    `json:"season"`
	Date        time.Time `json:"date"`
	HomeTeam    string    `json:"home_team"`
	AwayTeam    string    `json:"away_team"`
	HomeGoals   int       `json:"home_goals"`
	AwayGoals   int       `json:"away_goals"`
	Result      string    `json:"result"`
	SourceFile  string    `json:"source_file"`
}

type Config struct {
	InitialRating float64 `json:"initial_rating"`
	KFactor       float64 `json:"k_factor"`
	HomeAdvantage float64 `json:"home_advantage"`
	DrawFactor    float64 `json:"draw_factor"`
	Scale         float64 `json:"scale"`
}

type Probabilities struct {
	HomeWin float64 `json:"home_win"`
	Draw    float64 `json:"draw"`
	AwayWin float64 `json:"away_win"`
}

type Metrics struct {
	Matches          int     `json:"matches"`
	Correct          int     `json:"correct"`
	Accuracy         float64 `json:"accuracy"`
	LogLoss          float64 `json:"log_loss"`
	BrierScore       float64 `json:"brier_score"`
	ActualHomeWin    float64 `json:"actual_home_win_rate"`
	ActualDraw       float64 `json:"actual_draw_rate"`
	ActualAwayWin    float64 `json:"actual_away_win_rate"`
	PredictedHomeWin float64 `json:"predicted_home_win_rate"`
	PredictedDraw    float64 `json:"predicted_draw_rate"`
	PredictedAwayWin float64 `json:"predicted_away_win_rate"`
}

type RunResult struct {
	Config        Config  `json:"config"`
	Metrics       Metrics `json:"metrics"`
	StartSeason   string  `json:"start_season"`
	EndSeason     string  `json:"end_season"`
	HoldoutSeason string  `json:"holdout_season,omitempty"`
}

type BacktestReport struct {
	Competition     string    `json:"competition"`
	TrainingSeasons []string  `json:"training_seasons"`
	HoldoutSeason   string    `json:"holdout_season"`
	ChosenConfig    Config    `json:"chosen_config"`
	Training        Metrics   `json:"training"`
	Holdout         Metrics   `json:"holdout"`
	GeneratedAt     time.Time `json:"generated_at"`
	SourceFiles     []string  `json:"source_files"`
}
