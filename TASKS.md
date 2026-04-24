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
- [x] SEC-009 Integrate the better draw-decay baseline into the main Go backtest path
- [x] SEC-009b Add ordered probit model family to the grid
- [x] SEC-010 Add a naive-frequency benchmark and measure delta over that baseline
- [x] SEC-011 Build cross-league validation framework (generic CSV loader, cross-league CLI, download script)

## Phase 3 — Player-level modelling foundation
- [x] SEC-012 Define player-level domain models (Player, Team, Lineup, MatchEvent, Goal, Shot, Card, Substitution, Assist)
- [x] SEC-013 Add player-level data ingestion contracts (PlayerMatchStats, CSV loader, test fixtures)
- [x] SEC-014 Build team-form feature extraction from result history
- [x] SEC-015 Add feature registry and feature flag infrastructure
- [ ] SEC-016 Integrate team-form features into match prediction pipeline
- [ ] SEC-017 Build match-event timeline reconstructor from result data

## Immediate next task
Download additional league data with `python3 calibration/data/download_leagues.py`, then run `go run ./cmd/simcli cross-league` to benchmark transferability.
