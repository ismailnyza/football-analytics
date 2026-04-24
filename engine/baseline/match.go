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
	ModelFamily   string  `json:"model_family"`
	InitialRating float64 `json:"initial_rating"`
	KFactor       float64 `json:"k_factor"`
	HomeAdvantage float64 `json:"home_advantage"`
	DrawFactor    float64 `json:"draw_factor,omitempty"`
	BaseDraw      float64 `json:"base_draw,omitempty"`
	DrawScale     float64 `json:"draw_scale,omitempty"`
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

type MatchPrediction struct {
	Date             time.Time     `json:"date"`
	Season           string        `json:"season"`
	HomeTeam         string        `json:"home_team"`
	AwayTeam         string        `json:"away_team"`
	ActualHomeGoals  int           `json:"actual_home_goals"`
	ActualAwayGoals  int           `json:"actual_away_goals"`
	ActualResult     string        `json:"actual_result"`
	PredictedResult  string        `json:"predicted_result"`
	Probabilities    Probabilities `json:"probabilities"`
	HomeRatingBefore float64       `json:"home_rating_before"`
	AwayRatingBefore float64       `json:"away_rating_before"`
	SourceFile       string        `json:"source_file"`
}

type BacktestReport struct {
	Competition      string            `json:"competition"`
	PretrainSeasons  []string          `json:"pretrain_seasons"`
	ValidationSeason string            `json:"validation_season"`
	HoldoutSeason    string            `json:"holdout_season"`
	ChosenConfig     Config            `json:"chosen_config"`
	Pretrain         Metrics           `json:"pretrain"`
	Validation       Metrics           `json:"validation"`
	Holdout          Metrics           `json:"holdout"`
	HoldoutMatches   []MatchPrediction `json:"holdout_matches"`
	GeneratedAt      time.Time         `json:"generated_at"`
	SourceFiles      []string          `json:"source_files"`
}
