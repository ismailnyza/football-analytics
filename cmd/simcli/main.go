package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ismailnyza/football-analytics/engine/baseline"
	"github.com/ismailnyza/football-analytics/engine/evidence"
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
