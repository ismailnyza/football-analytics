package predict

import (
	"fmt"
	"sort"

	"github.com/ismailnyza/football-analytics/engine/baseline"
)

type OptimizerConfig struct {
	MaxIterations  int
	TargetAccuracy float64
	LearningRate   float64
	MinImprovement float64
	Verbose        bool
}

type OptimizationResult struct {
	Iterations      int
	BestAccuracy    float64
	BestConfig      baseline.Config
	AccuracyHistory []float64
	Converged       bool
	ErrorMatrix     map[string]int
}

type seasonalResult struct {
	accuracy float64
	config   baseline.Config
	errors   map[string]int
}

func OptimizeSeason(pretrain, validation, holdout []baseline.Match, initialCfg baseline.Config, optCfg OptimizerConfig) OptimizationResult {
	bestCfg := initialCfg
	_, _, _, ratings, _ := baseline.Tune(pretrain, validation, []baseline.Config{initialCfg})
	bestResult := runSeasonWithConfig(holdout, pretrain, validation, ratings, initialCfg)
	bestAccuracy := bestResult.accuracy

	history := []float64{bestAccuracy}
	if optCfg.Verbose {
		fmt.Printf("Iter 0: accuracy=%.1f%% config=(HA=%.0f K=%.0f BD=%.2f DS=%.0f)\n",
			bestAccuracy*100, bestCfg.HomeAdvantage, bestCfg.KFactor, bestCfg.BaseDraw, bestCfg.DrawScale)
	}

	params := []struct {
		name     string
		get      func(baseline.Config) float64
		set      func(*baseline.Config, float64)
		step     float64
		min, max float64
	}{
		{"home_adv", func(c baseline.Config) float64 { return c.HomeAdvantage }, func(c *baseline.Config, v float64) { c.HomeAdvantage = v }, 5, 30, 120},
		{"k_factor", func(c baseline.Config) float64 { return c.KFactor }, func(c *baseline.Config, v float64) { c.KFactor = v }, 2, 12, 48},
		{"base_draw", func(c baseline.Config) float64 { return c.BaseDraw }, func(c *baseline.Config, v float64) { c.BaseDraw = v }, 0.02, 0.20, 0.60},
		{"draw_scale", func(c baseline.Config) float64 { return c.DrawScale }, func(c *baseline.Config, v float64) { c.DrawScale = v }, 5, 30, 250},
	}

	for iter := 1; iter <= optCfg.MaxIterations; iter++ {
		improved := false

		for _, param := range params {
			testCfg := bestCfg

			val := param.get(testCfg) + param.step
			if val <= param.max {
				param.set(&testCfg, val)
				result := runSeasonWithConfig(holdout, pretrain, validation, ratings, testCfg)
				if result.accuracy > bestAccuracy+optCfg.MinImprovement {
					bestCfg = testCfg
					bestAccuracy = result.accuracy
					improved = true
					if optCfg.Verbose {
						fmt.Printf("Iter %d: +%s accuracy=%.1f%% (%s=%.1f)\n",
							iter, param.name, bestAccuracy*100, param.name, val)
					}
				}
			}

			testCfg = bestCfg
			val = param.get(testCfg) - param.step
			if val >= param.min {
				param.set(&testCfg, val)
				result := runSeasonWithConfig(holdout, pretrain, validation, ratings, testCfg)
				if result.accuracy > bestAccuracy+optCfg.MinImprovement {
					bestCfg = testCfg
					bestAccuracy = result.accuracy
					improved = true
					if optCfg.Verbose {
						fmt.Printf("Iter %d: -%s accuracy=%.1f%% (%s=%.1f)\n",
							iter, param.name, bestAccuracy*100, param.name, val)
					}
				}
			}
		}

		history = append(history, bestAccuracy)
		if bestAccuracy >= optCfg.TargetAccuracy {
			if optCfg.Verbose {
				fmt.Printf("Target accuracy %.0f%% reached at iter %d\n", optCfg.TargetAccuracy*100, iter)
			}
			return OptimizationResult{
				Iterations:      iter,
				BestAccuracy:    bestAccuracy,
				BestConfig:      bestCfg,
				AccuracyHistory: history,
				Converged:       true,
			}
		}

		if !improved {
			if optCfg.Verbose {
				fmt.Printf("Converged at iter %d with accuracy=%.1f%%\n", iter, bestAccuracy*100)
			}
			break
		}
	}

	return OptimizationResult{
		Iterations:      len(history) - 1,
		BestAccuracy:    bestAccuracy,
		BestConfig:      bestCfg,
		AccuracyHistory: history,
		Converged:       bestAccuracy >= optCfg.TargetAccuracy,
	}
}

func runSeasonWithConfig(holdout, pretrain, validation []baseline.Match, initialRatings map[string]float64, cfg baseline.Config) seasonalResult {
	ratings := cloneMap(initialRatings)
	form := baseline.ExtractTeamForm(append(pretrain, validation...), 5)
	pool := NewPlayerPool(ratings)
	pipeline := NewPipelineWithPool(pool, ratings, form, cfg)

	sorted := make([]baseline.Match, len(holdout))
	copy(sorted, holdout)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Date.Before(sorted[j].Date) })

	errMatrix := make(map[string]int)
	correct := 0
	total := 0

	for _, m := range sorted {
		total++
		pred, err := pipeline.Predict(m.HomeTeam, m.AwayTeam)
		if err != nil {
			continue
		}
		if pred.PredictedResult == m.Result {
			correct++
		} else {
			key := fmt.Sprintf("pred_%s_actual_%s", pred.PredictedResult, m.Result)
			errMatrix[key]++
		}

		actualHomeScore := 0.0
		switch m.Result {
		case "H":
			actualHomeScore = 1
		case "D":
			actualHomeScore = 0.5
		}
		expectedHomeScore := pred.ResultProbabilities.HomeWin + 0.5*pred.ResultProbabilities.Draw
		delta := cfg.KFactor * (actualHomeScore - expectedHomeScore)
		ratings[m.HomeTeam] = ratings[m.HomeTeam] + delta
		ratings[m.AwayTeam] = ratings[m.AwayTeam] - delta

		history := append(append([]baseline.Match{}, pretrain...), validation...)
		history = append(history, sorted[:total]...)
		form = baseline.ExtractTeamForm(history, 5)
		pipeline = NewPipelineWithPool(pool, ratings, form, cfg)
	}

	return seasonalResult{
		accuracy: float64(correct) / float64(total),
		config:   cfg,
		errors:   errMatrix,
	}
}

func cloneMap(src map[string]float64) map[string]float64 {
	dst := make(map[string]float64)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
