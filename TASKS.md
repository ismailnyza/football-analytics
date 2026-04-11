# Football Simulation Engine Task Queue

This file tracks the ordered execution plan.

Status values:
- NOT_STARTED
- IN_PROGRESS
- BLOCKED
- IMPLEMENTED
- REVIEWED
- TESTED
- DONE

---

## Phase A: Foundation

- [x] SEC-001 Project scaffolding and repo layout
- [x] SEC-002 Go module, build, and CLI bootstrap
- [x] SEC-003 SQLite schema and migrations
- [x] SEC-004 Domain models
- [x] SEC-005 Repository layer
- [x] SEC-007 TUI app shell and navigation

---

## Phase B: Core Match Simulation

- [x] SEC-020 Lineup selection engine
- [x] SEC-021 Match engine tick loop
- [x] SEC-022 Match action resolution
- [x] SEC-023 Fatigue system
- [x] SEC-024 Injury engine
- [x] SEC-025 Cards and suspension system
- [x] SEC-026 Set piece engine
- [x] SEC-027 Match persistence and stats aggregation
- [x] SEC-011 Match Lab screen
- [x] SEC-012 Match result and event log screens

---

## Phase C: Season Simulation

- [x] SEC-028 Fixture generation
- [x] SEC-029 League table engine
- [x] SEC-030 Season progression loop
- [x] SEC-031 Recovery and between-match updates
- [x] SEC-013 Season Lab screen
- [x] SEC-008 Dashboard screen
- [x] SEC-009 Squad screen
- [x] SEC-010 Player detail screen
- [x] SEC-014 Club screen

---

## Phase D: Long-Term World Simulation

- [x] SEC-032 Player development engine
- [x] SEC-033 Retirement and aging system
- [x] SEC-034 Youth intake / regen system
- [x] SEC-035 Club finance engine
- [x] SEC-036 Transfer valuation and decision engine
- [x] SEC-037 Contract logic
- [x] SEC-015 Transfers screen
- [x] SEC-017 World screen
- [x] SEC-016 Scenarios / branch management screen

---

## Phase E: Data and Calibration Layer

- [x] SEC-038 Ingestion raw staging pipeline
- [x] SEC-039 Source adapters
- [x] SEC-040 Normalization pipeline
- [x] SEC-041 Entity resolution
- [x] SEC-042 Validation and publish pipeline
- [x] SEC-018 Data / import screen
- [x] SEC-043 Python calibration bridge
- [x] SEC-044 Export tools
- [x] SEC-045 Test harness and realism regression suite
- [x] SEC-046 Documentation and handoff hygiene

---

## Phase F: Player Scraper

- [x] SEC-047 FBref HTML scraper adapter (player stats → staged records)
- [x] SEC-048 FBref normalizer (raw stats → domain player approximation)
- [x] SEC-049 CLI scrape command (`simcli scrape`)
- [x] SEC-050 Data/Import screen scrape trigger and status (`s` + `fbref_url.txt`, ledger updates)

---

## Immediate next task

Build on the SQLite-backed runtime with richer world workflows:
- add TUI browsing for published entities beyond the current normalized/raw previews
- expose branch creation and replay workflows more broadly across saved history outside Match/Season Lab
- consider persisting richer match summary detail for higher-fidelity replay views

---

## Notes

- `SEC-001` and `SEC-002` are complete only as a minimal bootstrap.
- `SEC-003` currently covers embedded SQL migrations and schema definition.
- `SEC-004` currently covers canonical entity structs and basic invariants.
- `SEC-005` currently covers repository interfaces and a SQLite SQL store layer with migration support and CRUD/list foundations.
- `SEC-007` now provides a Bubble Tea app shell with top bar, navigation, main content area, and status/help bar.
- `SEC-020` now provides deterministic formation templates and lineup selection with locked-player support.
- `SEC-021` now provides a deterministic 900-tick match loop skeleton with possession and chance generation.
- `SEC-022` now resolves attacking phases into build-up, penetration, shots, saves, blocks, goals, and turnovers.
- `SEC-023` now applies deterministic per-lineup fatigue accumulation and strength degradation across ticks.
- `SEC-024` now applies deterministic fatigue-driven injury triggering and injury events during matches.
- `SEC-025` now applies deterministic yellow/red card events and immediate red-card suspensions during matches.
- `SEC-026` now routes some advanced attacks into deterministic corners, free kicks, and penalties.
- `SEC-027` now aggregates stable match stats and provides a service-level persistence path for match records and event logs.
- `SEC-028–031` implement round-robin fixture generation, league table computation, full-season progression loop, and fatigue/injury recovery.
- `SEC-032–034` implement player development (growth/decline), retirement/aging, and youth intake with deterministic generation.
- `SEC-035–037` implement club finances (revenue/wage/budget), transfer valuation/AI decisions, and contract logic.
- `SEC-038–042` implement the data ingestion pipeline: staging store, source adapters, normalization, entity resolution, and validation.
- Ingest ledger `ingest_state.json` under `ResolveStateDir()` (default user cache `football-analytics/`) stores per-source last fetch time, counts, `cap_per_run`, and last error; `simcli fetch` and Data/Import `f` run demo adapters with capped staging.
- `simcli scrape --url …` or `fbref_url.txt` in state dir: HTTP fetch FBref `stats_table` player rows → staging; `NormalizeFbrefPlayerRecord` maps raw JSON to `NormalizedRecord` / validation attributes; Data/Import `s` runs the same scrape when `fbref_url.txt` exists.
- `SEC-043` provides a file-based Python calibration bridge (request/response JSON I/O).
- `SEC-044` provides CSV and JSON export tools for players, standings, and match results.
- `SEC-045` provides a realism regression suite verifying structural integrity and determinism of the full season simulation.
- `simtui` now bootstraps a local SQLite database under the resolved state dir, migrates it on startup, seeds a default demo world if empty, and feeds storage-backed data into Dashboard, World, Club, Squad, and Player Detail screens.
- `Match Lab` and `Season Lab` now load clubs/squads from the active SQLite branch when available and persist simulated seasons, fixtures, matches, and event logs back into SQLite.
- Data / Import now supports editing and saving the FBref scrape URL directly inside the TUI (`u` to edit, `enter` to save, `s` to scrape).
- Data ingest staging now persists in SQLite raw payload tables instead of process-local memory, and Match/Season Lab both expose saved-history reload flows from persisted SQLite data.
- `Scenarios` now updates shared active-branch runtime state, Data / Import shows staged raw payload previews per source, and Match Lab can create a new replay branch from a saved match.
- Data / Import now renders normalized FBref preview rows from staged payloads, Season Lab can create replay branches from saved seasons, and replay reconstruction now restores cards/injuries/suspensions from persisted match events.
- Always update `docs/IMPLEMENTATION_TRACKER.md` when task status changes.
