package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ismailnyza/football-analytics/engine/baseline"
	"github.com/ismailnyza/football-analytics/engine/evidence"
	"github.com/ismailnyza/football-analytics/engine/predict"
)

func main() {
	command := "status"
	if len(os.Args) >= 2 {
		command = os.Args[1]
	}

	switch command {
	case "status":
		printStatus()
	case "baseline-backtest":
		if err := runBaselineBacktest(); err != nil {
			fmt.Fprintf(os.Stderr, "baseline-backtest failed: %v\n", err)
			os.Exit(1)
		}
	case "cross-league":
		if err := runCrossLeague(); err != nil {
			fmt.Fprintf(os.Stderr, "cross-league failed: %v\n", err)
			os.Exit(1)
		}
	case "naive-frequency":
		if err := runNaiveFrequency(); err != nil {
			fmt.Fprintf(os.Stderr, "naive-frequency failed: %v\n", err)
			os.Exit(1)
		}
	case "predict":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "usage: simcli predict <HomeTeam> <AwayTeam>\n")
			os.Exit(1)
		}
		if err := runPredict(os.Args[2], os.Args[3]); err != nil {
			fmt.Fprintf(os.Stderr, "predict failed: %v\n", err)
			os.Exit(1)
		}
	case "predict-season":
		if err := runPredictSeason(); err != nil {
			fmt.Fprintf(os.Stderr, "predict-season failed: %v\n", err)
			os.Exit(1)
		}
	case "optimize":
		if err := runOptimize(); err != nil {
			fmt.Fprintf(os.Stderr, "optimize failed: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		os.Exit(1)
	}
}

func printStatus() {
	factor := evidence.FactorRecord{
		Name:       "historical team strength baseline (draw-decay)",
		Hypothesis: "stronger teams with explicit draw decay improve match-outcome prediction over naive priors",
		Grade:      evidence.GradeB,
		TestMethod: "deterministic chronological Go backtest on 2019/20–2024/25 EPL with 2024/25 holdout",
	}

	fmt.Println("self-improving football simulation research repo")
	fmt.Println("status: iteration 3 — draw-decay baseline integrated; holdout accuracy 53.42%")
	fmt.Printf("committed factor: %s (%s)\n", factor.Name, factor.Grade)
}

func runBaselineBacktest() error {
	matches, sourceFiles, err := baseline.LoadMatchesFromGlob(filepath.Join("data", "raw", "football-data", "E0_*.csv"))
	if err != nil {
		return err
	}
	report, err := baseline.RunBacktest(matches, "2324", "2425", baseline.DefaultGrid(), sourceFiles)
	if err != nil {
		return err
	}
	reportPath := filepath.Join("docs", "validation", "iteration-003-baseline-backtest.json")
	predictionsPath := filepath.Join("docs", "validation", "iteration-003-holdout-predictions.jsonl")
	if err := baseline.WriteReport(reportPath, report); err != nil {
		return err
	}
	if err := baseline.WritePredictionsJSONL(predictionsPath, report.HoldoutMatches); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", reportPath)
	fmt.Printf("wrote %s\n", predictionsPath)
	fmt.Printf("validation accuracy: %.4f\n", report.Validation.Accuracy)
	fmt.Printf("holdout accuracy: %.4f\n", report.Holdout.Accuracy)
	fmt.Printf("holdout predicted draw rate: %.4f\n", report.Holdout.PredictedDraw)
	fmt.Printf("model=%s K=%.0f home_adv=%.0f draw_factor=%.2f base_draw=%.2f draw_scale=%.0f draw_cut=%.2f\n",
		report.ChosenConfig.ModelFamily,
		report.ChosenConfig.KFactor,
		report.ChosenConfig.HomeAdvantage,
		report.ChosenConfig.DrawFactor,
		report.ChosenConfig.BaseDraw,
		report.ChosenConfig.DrawScale,
		report.ChosenConfig.DrawCut,
	)
	return nil
}

func runCrossLeague() error {
	type leagueEntry struct {
		Code    string
		Pattern string
	}
	leagues := []leagueEntry{
		{Code: "EPL", Pattern: "E0"},
	}
	for i := range leagues {
		l := &leagues[i]
		fmt.Printf("--- %s (%s) ---\n", l.Code, l.Pattern)
		matches, sourceFiles, err := baseline.LoadMatchesFromGlob(filepath.Join("data", "raw", "football-data", l.Pattern+"_*.csv"))
		if err != nil {
			fmt.Printf("  SKIP: %v\n", err)
			continue
		}
		report, err := baseline.RunBacktest(matches, "2324", "2425", baseline.DefaultGrid(), sourceFiles)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			continue
		}
		fmt.Printf("  validation accuracy: %.4f\n", report.Validation.Accuracy)
		fmt.Printf("  holdout accuracy:    %.4f\n", report.Holdout.Accuracy)
		fmt.Printf("  holdout draw rate:   %.4f\n", report.Holdout.PredictedDraw)
		fmt.Printf("  model: %s K=%.0f HA=%.0f\n", report.ChosenConfig.ModelFamily, report.ChosenConfig.KFactor, report.ChosenConfig.HomeAdvantage)
	}
	return nil
}

func runPredict(homeTeam, awayTeam string) error {
	matches, _, err := baseline.LoadMatchesFromGlob(filepath.Join("data", "raw", "football-data", "E0_*.csv"))
	if err != nil {
		return err
	}
	pretrain, validation, _ := baseline.SplitBySeason(matches, "2324", "2425")
	_, _, _, ratings, err := baseline.Tune(pretrain, validation, baseline.DefaultGrid())
	if err != nil {
		return err
	}
	form := baseline.ExtractTeamForm(append(pretrain, validation...), 5)
	cfg := baseline.Config{
		ModelFamily:   baseline.ModelFamilyDrawDecay,
		InitialRating: 1500,
		KFactor:       30,
		HomeAdvantage: 75,
		BaseDraw:      0.38,
		DrawScale:     75,
		Scale:         400,
	}
	playerPool, err := predict.LoadPlayerPoolFromJSON(filepath.Join("data", "players", "epl_2024_2025.json"))
	if err != nil {
		playerPool = predict.NewPlayerPool(ratings)
	}
	pipeline := predict.NewPipelineWithPool(playerPool, ratings, form, cfg)
	pred, err := pipeline.Predict(homeTeam, awayTeam)
	if err != nil {
		return err
	}
	data, _ := json.MarshalIndent(pred, "", "  ")
	fmt.Println(string(data))
	return nil
}

func runPredictSeason() error {
	matches, _, err := baseline.LoadMatchesFromGlob(filepath.Join("data", "raw", "football-data", "E0_*.csv"))
	if err != nil {
		return err
	}
	pretrain, validation, holdout := baseline.SplitBySeason(matches, "2324", "2425")
	_, _, _, ratings, err := baseline.Tune(pretrain, validation, baseline.DefaultGrid())
	if err != nil {
		return err
	}

	sort.Slice(holdout, func(i, j int) bool { return holdout[i].Date.Before(holdout[j].Date) })

	cfg := baseline.Config{
		ModelFamily:   baseline.ModelFamilyDrawDecay,
		InitialRating: 1500,
		KFactor:       30,
		HomeAdvantage: 75,
		BaseDraw:      0.38,
		DrawScale:     75,
		Scale:         400,
	}

	form := baseline.ExtractTeamForm(append(pretrain, validation...), 5)
	playerPool, err := predict.LoadPlayerPoolFromJSON(filepath.Join("data", "players", "epl_2024_2025.json"))
	if err != nil {
		playerPool = predict.NewPlayerPool(ratings)
	}
	pipeline := predict.NewPipelineWithPool(playerPool, ratings, form, cfg)

	var correctResult, correctScoreline int
	total := 0

	fmt.Println(strings.Repeat("=", 120))
	fmt.Printf("%-4s %-12s %-18s %-18s %-3s %-3s %-5s %-5s %-4s %-20s\n",
		"#", "Date", "Home", "Away", "Pr", "Ac", "PrSc", "AcSc", "Ok?", "TopScorer(Pr)")
	fmt.Println(strings.Repeat("-", 120))

	for i, m := range holdout {
		total++
		pred, err := pipeline.Predict(m.HomeTeam, m.AwayTeam)
		if err != nil {
			fmt.Fprintf(os.Stderr, "predict %s vs %s: %v\n", m.HomeTeam, m.AwayTeam, err)
			continue
		}

		resultOk := pred.PredictedResult == m.Result
		actualScore := fmt.Sprintf("%d-%d", m.HomeGoals, m.AwayGoals)
		scoreOk := pred.PredictedScoreline == actualScore

		if resultOk {
			correctResult++
		}
		if scoreOk {
			correctScoreline++
		}

		topScorer := ""
		if len(pred.HomeScorers) > 0 && len(pred.AwayScorers) > 0 {
			if pred.PredictedResult == "H" {
				topScorer = pred.HomeScorers[0].Player.Name
			} else if pred.PredictedResult == "A" {
				topScorer = pred.AwayScorers[0].Player.Name
			} else {
				topScorer = pred.HomeScorers[0].Player.Name + "/" + pred.AwayScorers[0].Player.Name
			}
		}

		resultMark := " "
		if resultOk {
			resultMark = "*"
		}

		fmt.Printf("%-4d %-12s %-18s %-18s %-3s %-3s %-5s %-5s %-4s %-20s\n",
			i+1,
			m.Date.Format("2006-01-02"),
			truncate(m.HomeTeam, 17),
			truncate(m.AwayTeam, 17),
			pred.PredictedResult,
			m.Result,
			pred.PredictedScoreline,
			actualScore,
			resultMark,
			truncate(topScorer, 19),
		)

		actualHomeScore := scoreForResult(m.Result)
		expectedHomeScore := pred.ResultProbabilities.HomeWin + 0.5*pred.ResultProbabilities.Draw
		delta := cfg.KFactor * (actualHomeScore - expectedHomeScore)
		ratings[m.HomeTeam] = ratings[m.HomeTeam] + delta
		ratings[m.AwayTeam] = ratings[m.AwayTeam] - delta

		history := append(append([]baseline.Match{}, pretrain...), validation...)
		history = append(history, holdout[:i+1]...)
		form = baseline.ExtractTeamForm(history, 5)
		pipeline = predict.NewPipelineWithPool(playerPool, ratings, form, cfg)
	}

	fmt.Println(strings.Repeat("-", 120))
	fmt.Printf("Result Accuracy: %d/%d = %.1f%%\n", correctResult, total, float64(correctResult)/float64(total)*100)
	fmt.Printf("Score Accuracy:  %d/%d = %.1f%%\n", correctScoreline, total, float64(correctScoreline)/float64(total)*100)
	fmt.Println(strings.Repeat("=", 120))
	return nil
}

func scoreForResult(result string) float64 {
	switch result {
	case "H":
		return 1
	case "D":
		return 0.5
	case "A":
		return 0
	}
	return 0.5
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "."
}

func runOptimize() error {
	matches, _, err := baseline.LoadMatchesFromGlob(filepath.Join("data", "raw", "football-data", "E0_*.csv"))
	if err != nil {
		return err
	}
	pretrain, validation, holdout := baseline.SplitBySeason(matches, "2324", "2425")

	initialCfg := baseline.Config{
		ModelFamily:   baseline.ModelFamilyDrawDecay,
		InitialRating: 1500,
		KFactor:       28,
		HomeAdvantage: 70,
		BaseDraw:      0.40,
		DrawScale:     75,
		Scale:         400,
	}

	optCfg := predict.OptimizerConfig{
		MaxIterations:  50,
		TargetAccuracy: 0.80,
		LearningRate:   0.1,
		MinImprovement: 0.002,
		Verbose:        true,
	}

	fmt.Println("Optimizing prediction parameters...")
	fmt.Println(strings.Repeat("-", 60))
	result := predict.OptimizeSeason(pretrain, validation, holdout, initialCfg, optCfg)

	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("Final accuracy:       %.1f%%\n", result.BestAccuracy*100)
	fmt.Printf("Iterations:           %d\n", result.Iterations)
	fmt.Printf("Converged to target:  %v\n", result.Converged)
	fmt.Printf("Best config: HA=%.0f K=%.0f BD=%.2f DS=%.0f\n",
		result.BestConfig.HomeAdvantage, result.BestConfig.KFactor,
		result.BestConfig.BaseDraw, result.BestConfig.DrawScale)
	return nil
}

func runNaiveFrequency() error {
	matches, sourceFiles, err := baseline.LoadMatchesFromGlob(filepath.Join("data", "raw", "football-data", "E0_*.csv"))
	if err != nil {
		return err
	}
	pretrain, _, holdout := baseline.SplitBySeason(matches, "2324", "2425")
	model := baseline.NewNaiveFrequency(pretrain)
	metrics := baseline.EvaluateNaiveFrequency(holdout, model)
	fmt.Printf("naive-frequency benchmark\n")
	fmt.Printf("pretrain outcome distribution: H=%.2f%% D=%.2f%% A=%.2f%%\n",
		model.HomeRate*100, model.DrawRate*100, model.AwayRate*100)
	fmt.Printf("always predicts: %s\n", model.Prediction)
	fmt.Printf("holdout accuracy: %.4f\n", metrics.Accuracy)
	fmt.Printf("holdout log loss: %.4f\n", metrics.LogLoss)
	fmt.Printf("holdout Brier: %.4f\n", metrics.BrierScore)
	fmt.Printf("source files: %v\n", sourceFiles)
	return nil
}
