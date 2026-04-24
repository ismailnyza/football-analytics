# Architecture

## System overview
The repository separates runtime simulation code, offline calibration work, and self-improvement infrastructure.

## Directory contract
- `cmd/` — entrypoints
- `engine/` — Go packages for evidence, simulation, and model logic
- `calibration/` — offline backtests and factor experiments
- `docs/` — theory, factors, validation, and research notes
- `infra/` — ledgers for prompt changes, self-audits, evidence, and disproven claims
- `data/` — datasets and provenance notes
- `video/` — video verification plans and annotations
- `.skills/` — repo-local workflows for future agent sessions

## Runtime ownership
Go owns
- deterministic runtime logic
- CLI entrypoints
- evidence-grade types used by the runtime

Python owns
- offline calibration
- batch analysis
- future notebooks and factor-testing scripts

## Initial runtime slice
The initial runtime slice is intentionally tiny:
- evidence grades are modeled in Go
- factor records can validate their own bookkeeping
- `simcli status` reports the bootstrap status

## First expansion path
1. add a historical-team-strength baseline contract
2. add a backtest runner under `calibration/backtest/`
3. add a first baseline predictor package under `engine/`
4. gate all future factor admission through the evidence ledger
