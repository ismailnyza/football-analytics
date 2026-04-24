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
Historical team-strength baseline for Premier League match result prediction is now implemented and measured.

Measured checkpoint
- training seasons: 2019/20 through 2023/24 EPL
- holdout season: 2024/25 EPL
- holdout exact W/D/L accuracy: 52.89%
- holdout log loss: 0.9946
- dominant residual: the baseline predicts zero draws despite a 24.47% actual draw rate in the holdout season

## Iteration 2 target
Reduce the draw-class error without leaking future information.

Deliverables for Iteration 2
- non-zero out-of-sample draw predictions
- improved holdout W/D/L accuracy over 52.89%
- explicit comparison against a naive-frequency benchmark
- updated evidence ledger and validation artifact

## Guardrails
- no factor enters the engine without a measured test
- no unsupported p-values or effect sizes in docs
- no runtime dependence on Python
- every infra update gets logged in `infra/`
