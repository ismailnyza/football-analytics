# Implementation Tracker

## Snapshot
- Branch state: dev
- Runtime state: Minimal Go CLI plus a measured EPL historical-team-strength baseline backtest
- Last verified commands:
  - `make fmt`
  - `go test ./...`
  - `go build ./...`
  - `go run ./cmd/simcli status`
  - `go run ./cmd/simcli baseline-backtest`

## Task ledger
| Task | Status | Notes |
| --- | --- | --- |
| SEC-000 Preserve git history and wipe the legacy working tree | TESTED | Removed previous working-tree contents while keeping `.git`, then rebuilt the repo from scratch. |
| SEC-001 Reinitialize repository structure | TESTED | Added clean top-level directories and foundational docs for engine, calibration, data, docs, infra, and video workflows. |
| SEC-002 Create infra ledgers | TESTED | Added `SKILL_REGISTRY.md`, `PROMPT_CHANGELOG.md`, `SELF_AUDIT.md`, `DISPROVEN_CEMETERY.md`, `EVIDENCE_MATRIX.md`, and `FACTOR_GRAPH.md`. |
| SEC-003 Minimal Go baseline | TESTED | Added evidence-grade primitives, factor validation rules, tests, and a `simcli status` command. |
| SEC-004 Initial research notes and repo-local skills | TESTED | Added iteration-0 audit/research notes and repo-local skills for agent infrastructure and evidence grading. |
| SEC-005 Historical-team-strength baseline contract | TESTED | Added `docs/research/historical-team-strength-baseline-contract.md` with input schema, modeling contract, evaluation contract, and acceptance criteria. |
| SEC-006 First backtest harness | TESTED | Added `engine/baseline` loader, Elo-with-draw baseline, tests, and `simcli baseline-backtest` to produce a deterministic validation artifact. |
| SEC-007 First validated baseline dataset | TESTED | Checked in six EPL season result CSVs under `data/raw/football-data/` and documented provenance under `data/provenance/football-data-premier-league.md`. |
| SEC-008 First evidence-graded factor result | TESTED | Ran the holdout backtest, wrote `docs/validation/iteration-001-baseline-backtest.json`, and recorded the measured result in `infra/EVIDENCE_MATRIX.md`. |

## Risks and caveats
- The current baseline achieves only 52.89% exact W/D/L accuracy on the 2024/25 EPL holdout, far below the 70% match-level target.
- The dominant failure mode is explicit: the current decision rule predicts zero draws out of sample while the holdout season draw rate is 24.47%.
- The evidence grade applies only to a single-league baseline predictor; there is no cross-league validation yet.
