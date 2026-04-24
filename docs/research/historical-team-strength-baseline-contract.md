# Historical Team-Strength Baseline Contract

## Goal
Produce the first measurable pre-match baseline for three-way match outcome prediction using only historical team strength inferred from prior results.

## Scope
This baseline is deliberately narrow.

Included
- competition: Premier League (`E0`)
- target: full-time match result (`H`, `D`, `A`)
- inputs available before kickoff only
- sequential chronological updates
- deterministic evaluation

Excluded
- player-level information
- injuries
- betting odds
- xG
- lineups
- transfers
- manager changes
- any factor not already encoded in the result history itself

## Input contract
Raw source
- files under `data/raw/football-data/E0_*.csv`
- source format: football-data.co.uk season CSVs

Required columns
- `Date`
- `HomeTeam`
- `AwayTeam`
- `FTHG`
- `FTAG`
- `FTR`

Parsed match schema
- `competition`: string
- `season`: string derived from filename
- `date`: parsed as `DD/MM/YYYY`
- `home_team`: string
- `away_team`: string
- `home_goals`: int
- `away_goals`: int
- `result`: one of `H`, `D`, `A`
- `source_file`: basename of the raw CSV

## Modeling contract
Baseline model
- Elo-style rolling team strength
- fixed home-advantage parameter in Elo points
- Davidson-style draw term for three-way probabilities
- updates occur only after each match is observed
- promoted or unseen clubs start at the initial rating

Parameter grid
- `k_factor`
- `home_advantage`
- `draw_factor`
- `scale`
- `initial_rating`

## Evaluation contract
Chronology
- training seasons: all seasons except the holdout season
- holdout season: currently `2425`
- no match may be predicted using future results

Primary metrics
- exact `W/D/L` accuracy
- multiclass log loss
- multiclass Brier score

Secondary diagnostics
- actual outcome distribution
- predicted class distribution

## Acceptance criteria for SEC-005
- raw data contract documented
- deterministic loader implemented and tested
- baseline model implemented and tested
- backtest command runs end to end on checked-in data
- first measured holdout accuracy written to `docs/validation/`
- first evidence entry recorded in `infra/EVIDENCE_MATRIX.md`

## Known limitations
- single-league baseline only
- no explicit handling of offseason squad turnover beyond rating carryover
- training metric is sequential one-step-ahead on the same seasons used for tuning, so it is not an unbiased generalization estimate
- holdout season is the first real out-of-sample checkpoint
