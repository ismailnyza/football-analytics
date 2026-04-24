package baseline

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"time"
)

func DefaultGrid() []Config {
	initial := 1500.0
	scale := 400.0
	kValues := []float64{12, 16, 20, 24, 28, 32}
	homeAdvValues := []float64{40, 50, 60, 70, 80, 90}
	drawValues := []float64{0.60, 0.75, 0.90, 1.05, 1.20}
	configs := make([]Config, 0, len(kValues)*len(homeAdvValues)*len(drawValues))
	for _, k := range kValues {
		for _, ha := range homeAdvValues {
			for _, draw := range drawValues {
				configs = append(configs, Config{
					InitialRating: initial,
					KFactor:       k,
					HomeAdvantage: ha,
					DrawFactor:    draw,
					Scale:         scale,
				})
			}
		}
	}
	return configs
}

func SplitBySeason(matches []Match, holdoutSeason string) (training []Match, holdout []Match) {
	for _, match := range matches {
		if match.Season == holdoutSeason {
			holdout = append(holdout, match)
		} else {
			training = append(training, match)
		}
	}
	return training, holdout
}

func Tune(training []Match, grid []Config) (Config, Metrics, error) {
	if len(training) == 0 {
		return Config{}, Metrics{}, fmt.Errorf("training matches required")
	}
	if len(grid) == 0 {
		return Config{}, Metrics{}, fmt.Errorf("config grid required")
	}

	best := grid[0]
	bestMetrics, _ := RunMatches(training, best, nil)
	for _, cfg := range grid[1:] {
		metrics, _ := RunMatches(training, cfg, nil)
		if better(metrics, bestMetrics) {
			best = cfg
			bestMetrics = metrics
		}
	}
	return best, bestMetrics, nil
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
	ratings := make(map[string]float64)
	for k, v := range initialRatings {
		ratings[k] = v
	}

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
	return metrics, ratings
}

func RunBacktest(matches []Match, holdoutSeason string, grid []Config, sourceFiles []string) (BacktestReport, error) {
	training, holdout := SplitBySeason(matches, holdoutSeason)
	if len(training) == 0 || len(holdout) == 0 {
		return BacktestReport{}, fmt.Errorf("need both training and holdout matches; got %d training and %d holdout", len(training), len(holdout))
	}
	cfg, trainingMetrics, err := Tune(training, grid)
	if err != nil {
		return BacktestReport{}, err
	}
	_, ratings := RunMatches(training, cfg, nil)
	holdoutMetrics, _ := RunMatches(holdout, cfg, ratings)
	trainingSeasons := uniqueSeasons(training)
	sort.Strings(trainingSeasons)
	return BacktestReport{
		Competition:     "EPL",
		TrainingSeasons: trainingSeasons,
		HoldoutSeason:   holdoutSeason,
		ChosenConfig:    cfg,
		Training:        trainingMetrics,
		Holdout:         holdoutMetrics,
		GeneratedAt:     time.Now().UTC(),
		SourceFiles:     sourceFiles,
	}, nil
}

func WriteReport(path string, report BacktestReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal backtest report: %w", err)
	}
	if err := os.WriteFile(path, append(data, byte('\n')), 0o644); err != nil {
		return fmt.Errorf("write backtest report: %w", err)
	}
	return nil
}

func predict(homeRating, awayRating float64, cfg Config) Probabilities {
	homeStrength := math.Pow(10, (homeRating+cfg.HomeAdvantage)/cfg.Scale)
	awayStrength := math.Pow(10, awayRating/cfg.Scale)
	drawStrength := cfg.DrawFactor * math.Sqrt(homeStrength*awayStrength)
	denom := homeStrength + awayStrength + drawStrength
	return Probabilities{
		HomeWin: homeStrength / denom,
		Draw:    drawStrength / denom,
		AwayWin: awayStrength / denom,
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

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
