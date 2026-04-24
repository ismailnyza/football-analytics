# Next Agent Playbook

## Quick-start commands
```bash
make fmt && make test
go run ./cmd/simcli status
go run ./cmd/simcli baseline-backtest
go run ./cmd/simcli naive-frequency
```

## Current state (Iteration 3 complete)
- Draw-decay baseline integrated: 53.42% holdout W/D/L accuracy on 2024/25 EPL
- Naive-frequency floor: 40.79%
- Delta over naive: +12.63%
- Player-level domain models defined in `engine/domain/`
- 15 tests passing across 3 packages

## Highest-priority next actions

### 1. Cross-league validation (SEC-011)
- Download La Liga, Serie A, Bundesliga, Ligue 1 data from football-data.co.uk
- Place in `data/raw/football-data/` with naming convention `E0_*.csv` → `SP1_*.csv`, `I1_*.csv`, `D1_*.csv`, `F1_*.csv`
- Run baseline backtest per league
- Record per-league metrics in evidence matrix
- Compare transferability: does a model tuned on EPL transfer to other leagues?

### 2. Improve draw prediction rate
- Current draw prediction: 3.42% vs actual 24.47%
- Search wider parameter space or add draw-bias term
- Consider ordered-probit / ordinal regression for three-outcome prediction

### 3. Player-level data ingestion (SEC-013)
- Source options: understat.com (xG), fbref.com (player stats), Opta via football-data
- Define data contract for player-level stats
- Create loader in `engine/domain/loader.go`
- Create test fixtures in `engine/domain/testdata/`

### 4. Team-form feature extraction (SEC-014)
- Extract rolling team form from existing result history
- Add form-weighted team strength to baseline model
- Measure accuracy delta

### 5. Feature registry (SEC-015)
- Create `engine/feature/` package
- Define `Feature` interface: Name, Compute(match) float64
- Feature flag system for toggle-able factors
- Integrate with evidence grading

## Rules reminder
- No factor enters the engine without a measured backtest
- Grade D/F factors stay out of runtime
- Python is offline-only
- Determinism first: same inputs = same outputs
- Update TASKS.md, PLANS.md, IMPLEMENTATION_TRACKER.md honestly
- Commit only passing work

## File locations reference
```
engine/domain/     — Player, Team, Lineup, MatchEvent types
engine/baseline/   — Elo model, backtest harness, naive-frequency
engine/evidence/   — Grade system (A-F), FactorRecord
cmd/simcli/        — CLI entrypoint
calibration/       — Python offline scripts
infra/             — Evidence matrix, factor graph, self-audit, prompt changelog
data/raw/          — Checked-in CSVs
docs/validation/   — Backtest artifact JSON files
```
