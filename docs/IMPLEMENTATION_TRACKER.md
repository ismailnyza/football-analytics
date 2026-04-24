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
| SEC-012 Player-level domain models | TESTED | Added `engine/domain/` package with Player, Team, Lineup, MatchEvent, Goal, Shot, Card, Substitution, and Assist types. 6 tests pass. |

## Risks and caveats
- Holdout accuracy improved from 52.89% (Davidson) to 53.42% (draw-decay) with non-zero draw predictions, but still below the offline best of 54.74%.
- The draw-decay model still under-predicts draws (3.42% vs actual 24.47%).
- Cross-league validation still does not exist.
- Player-level models are defined but have no data ingestion, feature extraction, or engine integration yet.
