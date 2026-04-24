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
| SEC-009 Draw-model search | TESTED | Added `calibration/backtest/search_draw_models.py` and `docs/validation/iteration-002-draw-search.json`; the best validation-supported draw-decay candidate reached 53.42% holdout accuracy with non-zero draw predictions. |

## Risks and caveats
- The committed Go baseline still achieves only 52.89% exact W/D/L accuracy on the 2024/25 EPL holdout, below both the research target and the stronger offline draw-decay candidates.
- The best offline holdout candidate reached 54.74% with non-zero draw predictions, but it is not yet the committed mainline baseline because the Go harness still uses the earlier Davidson family.
- Cross-league validation still does not exist.
