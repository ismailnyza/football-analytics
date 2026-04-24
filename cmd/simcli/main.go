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
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		os.Exit(1)
	}
}

func printStatus() {
	factor := evidence.FactorRecord{
		Name:       "historical team strength baseline",
		Hypothesis: "stronger teams should improve outcome prediction over naive priors",
		Grade:      evidence.GradeD,
		TestMethod: "planned backtest",
	}

	fmt.Println("self-improving football simulation research repo")
	fmt.Println("status: bootstrap complete, empirical baseline pending")
	fmt.Printf("next factor: %s (%s)\n", factor.Name, factor.Grade)
}

func runBaselineBacktest() error {
	matches, sourceFiles, err := baseline.LoadMatchesFromGlob(filepath.Join("data", "raw", "football-data", "E0_*.csv"))
	if err != nil {
		return err
	}
	report, err := baseline.RunBacktest(matches, "2425", baseline.DefaultGrid(), sourceFiles)
	if err != nil {
		return err
	}
	outPath := filepath.Join("docs", "validation", "iteration-001-baseline-backtest.json")
	if err := baseline.WriteReport(outPath, report); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", outPath)
	fmt.Printf("holdout accuracy: %.4f\n", report.Holdout.Accuracy)
	fmt.Printf("holdout log loss: %.4f\n", report.Holdout.LogLoss)
	fmt.Printf("holdout brier: %.4f\n", report.Holdout.BrierScore)
	fmt.Printf("config: K=%.0f home_adv=%.0f draw=%.2f scale=%.0f\n", report.ChosenConfig.KFactor, report.ChosenConfig.HomeAdvantage, report.ChosenConfig.DrawFactor, report.ChosenConfig.Scale)
	return nil
}
