# Data Provenance — football-data.co.uk Premier League results

## Source
- Provider: football-data.co.uk
- Competition code: `E0`
- Download pattern: `https://www.football-data.co.uk/mmz4281/<season>/E0.csv`

## Checked-in seasons
- `1920` → `data/raw/football-data/E0_1920.csv`
- `2021` → `data/raw/football-data/E0_2021.csv`
- `2122` → `data/raw/football-data/E0_2122.csv`
- `2223` → `data/raw/football-data/E0_2223.csv`
- `2324` → `data/raw/football-data/E0_2324.csv`
- `2425` → `data/raw/football-data/E0_2425.csv`

## Retrieval date
- 2026-04-25 UTC

## Purpose
These files provide the first empirical match-result history used to construct and evaluate the historical team-strength baseline.

## Caveats
- this source gives result history, not player-level tracking
- club names are source-native and may change slightly across seasons
- postponed/rescheduled fixtures are evaluated in the order provided after date parsing
