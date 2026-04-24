package baseline

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"time"
)

const (
	ModelFamilyDavidson  = "davidson"
	ModelFamilyDrawDecay = "draw_decay"
)

func DefaultGrid() []Config {
	initial := 1500.0
	scale := 400.0
	kValues := []float64{16, 20, 24, 28, 32}
	homeAdvValues := []float64{50, 60, 70, 80, 90}
	var configs []Config
	for _, k := range kValues {
		for _, ha := range homeAdvValues {
			for _, draw := range []float64{0.4, 0.6, 0.8, 1.0, 1.2} {
				configs = append(configs, Config{
					ModelFamily:   ModelFamilyDavidson,
					InitialRating: initial,
					KFactor:       k,
					HomeAdvantage: ha,
					DrawFactor:    draw,
					Scale:         scale,
				})
			}
			for _, baseDraw := range []float64{0.20, 0.25, 0.30, 0.35, 0.40} {
				for _, drawScale := range []float64{50, 75, 100, 125, 150} {
					configs = append(configs, Config{
						ModelFamily:   ModelFamilyDrawDecay,
						InitialRating: initial,
						KFactor:       k,
						HomeAdvantage: ha,
						BaseDraw:      baseDraw,
						DrawScale:     drawScale,
						Scale:         scale,
					})
				}
			}
		}
	}
	return configs
}

func SplitBySeason(matches []Match, validationSeason, holdoutSeason string) (pretrain, validation, holdout []Match) {
	for _, match := range matches {
		switch match.Season {
		case holdoutSeason:
			holdout = append(holdout, match)
		case validationSeason:
			validation = append(validation, match)
		default:
			pretrain = append(pretrain, match)
		}
	}
	return pretrain, validation, holdout
}

func Tune(pretrain, validation []Match, grid []Config) (Config, Metrics, Metrics, map[string]float64, error) {
	if len(pretrain) == 0 {
		return Config{}, Metrics{}, Metrics{}, nil, fmt.Errorf("pretrain matches required")
	}
	if len(validation) == 0 {
		return Config{}, Metrics{}, Metrics{}, nil, fmt.Errorf("validation matches required")
	}
	if len(grid) == 0 {
		return Config{}, Metrics{}, Metrics{}, nil, fmt.Errorf("config grid required")
	}

	best := grid[0]
	bestPretrain, ratings, _ := RunMatchesDetailed(pretrain, best, nil, false)
	bestValidation, bestPostValidationRatings, _ := RunMatchesDetailed(validation, best, ratings, false)
	bestRatings := cloneRatings(bestPostValidationRatings)

	for _, cfg := range grid[1:] {
		pretrainMetrics, cfgRatings, _ := RunMatchesDetailed(pretrain, cfg, nil, false)
		validationMetrics, cfgPostValidationRatings, _ := RunMatchesDetailed(validation, cfg, cfgRatings, false)
		if better(validationMetrics, bestValidation) {
			best = cfg
			bestPretrain = pretrainMetrics
			bestValidation = validationMetrics
			bestRatings = cloneRatings(cfgPostValidationRatings)
		}
	}
	return best, bestPretrain, bestValidation, bestRatings, nil
}

func better(candidate, incumbent Metrics) bool {
	if candidate.Accuracy > incumbent.Accuracy+1e-9 {
		return true
	}
	if math.Abs(candidate.Accuracy-incumbent.Accuracy) <= 1e-9 && candidate.LogLoss < incumbent.LogLoss {
		return true
	}
	return false
}

func RunMatches(matches []Match, cfg Config, initialRatings map[string]float64) (Metrics, map[string]float64) {
	metrics, ratings, _ := RunMatchesDetailed(matches, cfg, initialRatings, false)
	return metrics, ratings
}

func RunMatchesDetailed(matches []Match, cfg Config, initialRatings map[string]float64, includePredictions bool) (Metrics, map[string]float64, []MatchPrediction) {
	ratings := cloneRatings(initialRatings)
	predictions := make([]MatchPrediction, 0, len(matches))

	var correct int
	var logLoss float64
	var brier float64
	var actualHome, actualDraw, actualAway float64
	var predictedHome, predictedDraw, predictedAway float64

	for _, match := range matches {
		homeRating := getRating(ratings, match.HomeTeam, cfg.InitialRating)
		awayRating := getRating(ratings, match.AwayTeam, cfg.InitialRating)
		probs := predict(homeRating, awayRating, cfg)
		prediction := argmax(probs)
		if prediction == match.Result {
			correct++
		}

		actualVec := actualVector(match.Result)
		probVec := []float64{probs.HomeWin, probs.Draw, probs.AwayWin}
		logLoss += -math.Log(max(probabilityForResult(probs, match.Result), 1e-12))
		for i := range actualVec {
			d := probVec[i] - actualVec[i]
			brier += d * d
		}

		switch match.Result {
		case "H":
			actualHome++
		case "D":
			actualDraw++
		case "A":
			actualAway++
		}
		switch prediction {
		case "H":
			predictedHome++
		case "D":
			predictedDraw++
		case "A":
			predictedAway++
		}

		if includePredictions {
			predictions = append(predictions, MatchPrediction{
				Date:             match.Date,
				Season:           match.Season,
				HomeTeam:         match.HomeTeam,
				AwayTeam:         match.AwayTeam,
				ActualHomeGoals:  match.HomeGoals,
				ActualAwayGoals:  match.AwayGoals,
				ActualResult:     match.Result,
				PredictedResult:  prediction,
				Probabilities:    probs,
				HomeRatingBefore: homeRating,
				AwayRatingBefore: awayRating,
				SourceFile:       match.SourceFile,
			})
		}

		expectedHomeScore := probs.HomeWin + 0.5*probs.Draw
		actualHomeScore := scoreForResult(match.Result)
		delta := cfg.KFactor * (actualHomeScore - expectedHomeScore)
		ratings[match.HomeTeam] = homeRating + delta
		ratings[match.AwayTeam] = awayRating - delta
	}

	total := float64(len(matches))
	metrics := Metrics{Matches: len(matches), Correct: correct}
	if total > 0 {
		metrics.Accuracy = float64(correct) / total
		metrics.LogLoss = logLoss / total
		metrics.BrierScore = brier / total
		metrics.ActualHomeWin = actualHome / total
		metrics.ActualDraw = actualDraw / total
		metrics.ActualAwayWin = actualAway / total
		metrics.PredictedHomeWin = predictedHome / total
		metrics.PredictedDraw = predictedDraw / total
		metrics.PredictedAwayWin = predictedAway / total
	}
	return metrics, ratings, predictions
}

func RunBacktest(matches []Match, validationSeason, holdoutSeason string, grid []Config, sourceFiles []string) (BacktestReport, error) {
	pretrain, validation, holdout := SplitBySeason(matches, validationSeason, holdoutSeason)
	if len(pretrain) == 0 || len(validation) == 0 || len(holdout) == 0 {
		return BacktestReport{}, fmt.Errorf("need pretrain, validation, and holdout matches; got %d, %d, %d", len(pretrain), len(validation), len(holdout))
	}
	cfg, pretrainMetrics, validationMetrics, ratings, err := Tune(pretrain, validation, grid)
	if err != nil {
		return BacktestReport{}, err
	}
	holdoutMetrics, _, holdoutPredictions := RunMatchesDetailed(holdout, cfg, ratings, true)
	pretrainSeasons := uniqueSeasons(pretrain)
	sort.Strings(pretrainSeasons)
	return BacktestReport{
		Competition:      "EPL",
		PretrainSeasons:  pretrainSeasons,
		ValidationSeason: validationSeason,
		HoldoutSeason:    holdoutSeason,
		ChosenConfig:     cfg,
		Pretrain:         pretrainMetrics,
		Validation:       validationMetrics,
		Holdout:          holdoutMetrics,
		HoldoutMatches:   holdoutPredictions,
		GeneratedAt:      time.Now().UTC(),
		SourceFiles:      sourceFiles,
	}, nil
}

func WriteReport(path string, report BacktestReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal backtest report: %w", err)
	}
	if err := os.WriteFile(path, append(data, byte(10)), 0o644); err != nil {
		return fmt.Errorf("write backtest report: %w", err)
	}
	return nil
}

func WritePredictionsJSONL(path string, predictions []MatchPrediction) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create predictions file: %w", err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	for _, prediction := range predictions {
		if err := encoder.Encode(prediction); err != nil {
			return fmt.Errorf("encode prediction: %w", err)
		}
	}
	return nil
}

func predict(homeRating, awayRating float64, cfg Config) Probabilities {
	switch cfg.ModelFamily {
	case ModelFamilyDrawDecay:
		return predictDrawDecay(homeRating, awayRating, cfg)
	case ModelFamilyDavidson, "":
		fallthrough
	default:
		return predictDavidson(homeRating, awayRating, cfg)
	}
}

func predictDavidson(homeRating, awayRating float64, cfg Config) Probabilities {
	homeStrength := math.Pow(10, (homeRating+cfg.HomeAdvantage)/cfg.Scale)
	awayStrength := math.Pow(10, awayRating/cfg.Scale)
	drawStrength := cfg.DrawFactor * math.Sqrt(homeStrength*awayStrength)
	denom := homeStrength + awayStrength + drawStrength
	return Probabilities{HomeWin: homeStrength / denom, Draw: drawStrength / denom, AwayWin: awayStrength / denom}
}

func predictDrawDecay(homeRating, awayRating float64, cfg Config) Probabilities {
	diff := (homeRating + cfg.HomeAdvantage) - awayRating
	pDraw := cfg.BaseDraw * math.Exp(-math.Abs(diff)/cfg.DrawScale)
	pDraw = clamp(pDraw, 0.01, 0.60)
	pHomeNoDraw := 1.0 / (1.0 + math.Pow(10, -diff/cfg.Scale))
	return Probabilities{
		Draw:    pDraw,
		HomeWin: (1.0 - pDraw) * pHomeNoDraw,
		AwayWin: (1.0 - pDraw) * (1.0 - pHomeNoDraw),
	}
}

func argmax(probs Probabilities) string {
	result := "H"
	best := probs.HomeWin
	if probs.Draw > best {
		result = "D"
		best = probs.Draw
	}
	if probs.AwayWin > best {
		result = "A"
	}
	return result
}

func probabilityForResult(probs Probabilities, result string) float64 {
	switch result {
	case "H":
		return probs.HomeWin
	case "D":
		return probs.Draw
	case "A":
		return probs.AwayWin
	default:
		return 0
	}
}

func actualVector(result string) []float64 {
	switch result {
	case "H":
		return []float64{1, 0, 0}
	case "D":
		return []float64{0, 1, 0}
	case "A":
		return []float64{0, 0, 1}
	default:
		return []float64{0, 0, 0}
	}
}

func scoreForResult(result string) float64 {
	switch result {
	case "H":
		return 1
	case "D":
		return 0.5
	case "A":
		return 0
	default:
		return 0.5
	}
}

func getRating(ratings map[string]float64, team string, initial float64) float64 {
	if rating, ok := ratings[team]; ok {
		return rating
	}
	ratings[team] = initial
	return initial
}

func uniqueSeasons(matches []Match) []string {
	seen := make(map[string]struct{})
	var seasons []string
	for _, match := range matches {
		if _, ok := seen[match.Season]; ok {
			continue
		}
		seen[match.Season] = struct{}{}
		seasons = append(seasons, match.Season)
	}
	return seasons
}

func cloneRatings(src map[string]float64) map[string]float64 {
	cloned := make(map[string]float64)
	for k, v := range src {
		cloned[k] = v
	}
	return cloned
}

func clamp(value, low, high float64) float64 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
