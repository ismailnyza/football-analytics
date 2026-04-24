# Task Queue

Status values
- NOT_STARTED
- IN_PROGRESS
- BLOCKED
- TESTED
- DONE

## Phase 0 — Bootstrap
- [x] SEC-000 Preserve git history and wipe the legacy working tree
- [x] SEC-001 Reinitialize repository structure for a self-improving research system
- [x] SEC-002 Create infra ledgers for prompt changes, self-audits, evidence, and disproven factors
- [x] SEC-003 Add a minimal buildable Go baseline with evidence-grade primitives and `simcli status`
- [x] SEC-004 Create initial research notes and repo-local skills for future iterations

## Phase 1 — First empirical baseline
- [x] SEC-005 Define the historical-team-strength baseline data contract
- [x] SEC-006 Build the first backtest harness for match result prediction
- [x] SEC-007 Ingest or check in the first validated baseline dataset with provenance notes
- [x] SEC-008 Record the first evidence-graded factor result in `infra/EVIDENCE_MATRIX.md`

## Phase 2 — Baseline error reduction
- [ ] SEC-009 Calibrate the draw mechanism so the baseline predicts non-zero draws out of sample
- [ ] SEC-010 Add a naive-frequency benchmark and measure delta over that baseline
- [ ] SEC-011 Expand the baseline across additional top-5 leagues and compare transferability

## Immediate next task
Implement SEC-009: improve the draw model so out-of-sample draw predictions are non-zero and holdout W/D/L accuracy exceeds the current 52.89% EPL benchmark.
