# Implementation Tracker

## Snapshot
- Branch state: dev
- Runtime state: Minimal Go CLI plus a measured EPL baseline and an offline draw-model search artifact
- Last verified commands:
  - `make fmt`
  - `go test ./...`
  - `go build ./...`
  - `go run ./cmd/simcli status`
  - `go run ./cmd/simcli baseline-backtest`
  - `python3 calibration/backtest/search_draw_models.py`

## Task ledger
| Task | Status | Notes |
| --- | --- | --- |
| SEC-000 Preserve git history and wipe the legacy working tree | TESTED | Removed previous working-tree contents while keeping `.git`, then rebuilt the repo from scratch. |
| SEC-001 Reinitialize repository structure | TESTED | Added clean top-level directories and foundational docs for engine, calibration, data, docs, infra, and video workflows. |
| SEC-002 Create infra ledgers | TESTED | Added `SKILL_REGISTRY.md`, `PROMPT_CHANGELOG.md`, `SELF_AUDIT.md`, `DISPROVEN_CEMETERY.md`, `EVIDENCE_MATRIX.md`, and `FACTOR_GRAPH.md`. |
| SEC-003 Minimal Go baseline | TESTED | Added evidence-grade primitives, factor validation rules, tests, and a `simcli status` command. |
| SEC-004 Initial research notes and repo-local skills | TESTED | Added iteration-0 audit/research notes and repo-local skills for agent infrastructure and evidence grading. |
| SEC-005 Historical-team-strength baseline contract | TESTED | Added `docs/research/historical-team-strength-baseline-contract.md` with input schema, modeling contract, evaluation contract, and acceptance criteria. |
| SEC-006 First backtest harness | TESTED | Added `engine/baseline` loader, Davidson-style Elo baseline, tests, and `simcli baseline-backtest` to produce a deterministic validation artifact. |
| SEC-007 First validated baseline dataset | TESTED | Checked in six EPL season result CSVs under `data/raw/football-data/` and documented provenance under `data/provenance/football-data-premier-league.md`. |
| SEC-008 First evidence-graded factor result | TESTED | Ran the holdout backtest, wrote `docs/validation/iteration-001-baseline-backtest.json`, and recorded the measured result in `infra/EVIDENCE_MATRIX.md`. |
| SEC-009 Draw-model search + Go integration | TESTED | The draw-decay baseline (k=28, HA=70, base_draw=0.40, draw_scale=75) is integrated into the main Go backtest path via DefaultGrid(). Holdout accuracy: 53.42% with 3.42% predicted draws. |
| SEC-010 Naive-frequency benchmark | TESTED | Added `engine/baseline/naive.go` with `NewNaiveFrequency` and `EvaluateNaiveFrequency`, plus `simcli naive-frequency` command. Floor accuracy: 40.79% (always predict Home). Delta over naive: +12.63%. |
| SEC-009b Ordered probit model | TESTED | Added `ModelFamilyOrderedProbit` with symmetric cut-point model. Produces 38-63% draw probability for even teams. Grid search still selects draw_decay on validation accuracy. |
| SEC-011 Cross-league framework | TESTED | Generic CSV loader auto-detects competition from filename. `simcli cross-league` command runs per-league backtests. Download script at `calibration/data/download_leagues.py`. |
| SEC-012 Player-level domain models | TESTED | Added `engine/domain/` package with Player, Team, Lineup, MatchEvent, Goal, Shot, Card, Substitution, and Assist types. 8 tests pass. |
| SEC-013 Player data ingestion | TESTED | Added `PlayerMatchStats` type, `LoadPlayerStatsFromGlob`, CSV parser, test fixtures in `engine/domain/testdata/`. |
| SEC-014 Team-form extraction | TESTED | Added `engine/baseline/form.go` with `ExtractTeamForm` and `CalculateFormDelta`. 4 form tests pass. |
| SEC-015 Feature registry | TESTED | Added `engine/feature/` package with `Registry`, `Feature` interface, `BaseFeature`. Grade-gated registration. 4 tests pass. |
| SEC-016 Match prediction pipeline | TESTED | Added `engine/predict/` package with `Pipeline`, `PlayerPool`, scorer/assist prediction, team event estimation, match timeline reconstruction. `simcli predict` command. 7 tests pass. |
| SEC-017 Match event reconstructor | TESTED | `ReconstructMatch` generates full event timeline: goals with scorer names and minute, assists, cards. Half-time score. Deterministic. |

## Risks and caveats
- Draw-decay still under-predicts draws (3.42% vs actual 24.47%). Ordered probit is coded but not selected by validation.
- Cross-league framework exists but only EPL data is checked in. Need to download other leagues.
- Player pool is synthetic (generated from team ratings). Real player data needed for accurate scorer/assist predictions.
- Match event timeline is plausible but generated, not from real data.
