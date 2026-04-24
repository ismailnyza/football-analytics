# Next Agent Playbook

## Quick-start commands
```bash
make fmt && make test
go run ./cmd/simcli status
go run ./cmd/simcli baseline-backtest
go run ./cmd/simcli naive-frequency
go run ./cmd/simcli cross-league
python3 calibration/data/download_leagues.py        # download other leagues
python3 calibration/backtest/search_draw_models.py  # offline draw search
```

## Current state (Iteration 4 complete)
- Draw-decay baseline integrated: 53.42% holdout W/D/L accuracy
- Naive-frequency floor: 40.79%
- Ordered probit model coded (38-63% draw rate) but not selected by grid validation
- Cross-league framework: generic CSV loader, competition auto-detection
- Player domain models: Player, Team, Lineup, MatchEvent, Goal, Shot, Card, Sub, Assist
- Player data loader: PlayerMatchStats CSV parser with test fixtures
- Team-form extraction: rolling form from result history
- Feature registry: grade-gated, Feature interface, BaseFeature
- 27 tests passing across 4 packages

## Highest-priority next actions

### 1. Download other league data
```bash
python3 calibration/data/download_leagues.py
go run ./cmd/simcli cross-league
```
This will benchmark 5 leagues and produce per-league accuracy metrics.

### 2. Integrate team-form features into prediction (SEC-016)
- Add form-weighted rating adjustment to the Elo model
- Measure delta on holdout accuracy
- Record evidence in matrix

### 3. Build match-event timeline reconstructor (SEC-017)
- Use existing team stats (shots, corners, cards from CSVs) to create event timelines
- Map to MatchEvent types in engine/domain/event.go
- This bridges the gap between team-level data and atomic match modelling

### 4. Source real player-level data
- understat.com for xG/shots per player
- fbref.com for comprehensive player stats
- Load via the existing PlayerMatchStats CSV parser

### 5. Build player contribution model
- Use PlayerMatchStats to estimate player strength
- Build team strength from player-level data
- Compare accuracy to Elo-only baseline

## Rules reminder
- No factor enters the engine without a measured backtest
- Grade D/F factors stay out of runtime
- Python is offline-only
- Determinism first: same inputs = same outputs
- Update TASKS.md, PLANS.md, IMPLEMENTATION_TRACKER.md honestly
- Commit only passing work

## File locations
```
engine/baseline/   — Elo models, backtest, naive-frequency, team form
engine/domain/     — Player, Team, Lineup, MatchEvent, PlayerMatchStats, loader
engine/feature/    — Feature registry, BaseFeature
engine/evidence/   — Grade system (A-F), FactorRecord
cmd/simcli/        — CLI: status, baseline-backtest, naive-frequency, cross-league
calibration/       — Python: draw search, league download
infra/             — Evidence matrix, factor graph, self-audit, prompt changelog
data/raw/          — Checked-in CSVs (EPL only)
docs/validation/   — Backtest artifact JSON files
```
