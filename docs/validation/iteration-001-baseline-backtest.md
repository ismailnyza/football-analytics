# Iteration 001 — EPL Historical Team-Strength Baseline

## Dataset
- competition: Premier League (`E0`)
- training seasons: `1920`, `2021`, `2122`, `2223`, `2324`
- holdout season: `2425`
- matches: 1,900 training, 380 holdout

## Model
- sequential Elo-style team-strength baseline
- fixed home advantage
- Davidson-style draw term
- parameter grid tuned on training-history exact W/D/L accuracy with log loss as a tie-breaker

Chosen parameters
- initial rating: 1500
- K-factor: 24
- home advantage: 70 Elo points
- draw factor: 0.60
- scale: 400

## Holdout metrics
- exact W/D/L accuracy: 52.89%
- log loss: 0.9946
- Brier score: 0.5951

## Distribution check
- actual holdout outcomes: 40.79% home win, 24.47% draw, 34.74% away win
- predicted classes: 66.05% home win, 0.00% draw, 33.95% away win

## Main conclusion
The repository now has a real empirical baseline, but it is still badly underfit on the draw class. The next iteration should target draw calibration explicitly before adding richer factors.
