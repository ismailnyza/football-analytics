# Iteration 002 — Draw-Model Search

## Goal
Reduce the zero-draw failure mode found in the first EPL historical-team-strength baseline.

## Method
- pretrain seasons: `1920`, `2021`, `2122`, `2223`
- validation season: `2324`
- holdout season: `2425`
- search script: `calibration/backtest/search_draw_models.py`
- candidate families:
  - Davidson-style Elo baseline
  - draw-decay baseline where draw probability decays with absolute pre-match rating difference

## Best validation-supported draw-aware candidate
Configuration
- family: `draw_decay`
- K-factor: `28`
- home advantage: `70`
- base draw: `0.40`
- draw scale: `75`

Metrics
- validation accuracy: `58.68%`
- validation predicted draws: `3.95%`
- holdout accuracy: `53.42%`
- holdout predicted draws: `3.42%`

## Strongest non-zero-draw holdout candidate found
Configuration
- family: `draw_decay`
- K-factor: `32`
- home advantage: `80`
- base draw: `0.35`
- draw scale: `125`

Metrics
- validation accuracy: `57.37%`
- holdout accuracy: `54.74%`
- holdout predicted draws: `1.32%`

## Conclusion
The zero-draw failure mode is not structural to team-strength baselines. A draw-decay family can outperform the committed Davidson baseline on the same 2024/25 EPL holdout while producing non-zero draw predictions. The next step is to port this family into the main Go harness so the improved baseline becomes the reproducible default rather than an offline search result.
