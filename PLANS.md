# Execution Plan — Self-Improving Football Simulation Research

## Objective
Boot a clean repository that can improve both the football simulation engine and the agent infrastructure that builds it.

## Iteration 0: bootstrap from zero
1. scaffold the repository
2. create the infra ledgers
3. add a minimal executable baseline
4. document the first research target
5. keep all evidence claims honest and sparse

## Iteration 1 result
Historical team-strength baseline for Premier League match result prediction is implemented and measured.

Committed checkpoint
- method: Davidson-style Elo baseline in Go
- holdout season: 2024/25 EPL
- exact W/D/L accuracy: 52.89%
- draw predictions: 0.00%

## Iteration 2 result
A reproducible draw-model search found a stronger candidate family.

Experimental checkpoint
- method: draw-decay team-strength model searched offline under `calibration/backtest/search_draw_models.py`
- validation season: 2023/24 EPL
- best validation candidate with non-zero draws: `k=28`, `home_advantage=70`, `base_draw=0.40`, `draw_scale=75`
- holdout exact W/D/L accuracy for that candidate: 53.42%
- stronger non-zero-draw holdout candidate also exists at 54.74% but with weaker validation support

## Iteration 3 result
The draw-decay baseline is integrated into the main Go backtest path. The naive-frequency floor benchmark is implemented and measured.

Committed checkpoint
- method: draw-decay Elo baseline in Go
- holdout season: 2024/25 EPL
- holdout W/D/L accuracy: 53.42% (up from 52.89%)
- holdout predicted draw rate: 3.42% (up from 0.00%)
- naive-frequency floor: 40.79% (always predict Home)
- delta over naive floor: +12.63%
- model: k=28, home_advantage=70, base_draw=0.40, draw_scale=75
- player-level domain models defined in engine/domain/

## Iteration 4 target
Cross-league validation: run the baseline on La Liga, Serie A, Bundesliga, and Ligue 1. Integrate team-form features and start match-event timeline reconstruction.

## Guardrails
- no factor enters the engine without a measured test
- no unsupported p-values or effect sizes in docs
- no runtime dependence on Python
- every infra update gets logged in `infra/`
