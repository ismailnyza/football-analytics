# Implementation Tracker

This file is the single truth ledger for implementation, review, and test status.

## Snapshot

- Branch state: dev
- Runtime state: Full Go simulation engine with TUI shell
- Last verified commands:
  - `go build ./...`
  - `go test ./...`
  - `go test ./internal/app ./internal/ingestion ./internal/tui/screens/dataimport ./internal/tui/screens/matchlab ./internal/tui/screens/seasonlab ./internal/tui/... ./cmd/simtui`

## Task ledger

| Task | Status | Notes |
| --- | --- | --- |
| SEC-001 Project scaffolding and repo layout | TESTED | Created target top-level structure and committed placeholder packages/files so the scaffold is explicit in-repo. |
| SEC-002 Go module, build, and CLI bootstrap | TESTED | Added `go.mod`, `Makefile`, `cmd/simtui`, and `cmd/simcli` with standard-library bootstrap commands. Cobra and Bubble Tea are still deferred. |
| SEC-003 SQLite schema and migrations | TESTED | Added embedded migration loader, initial schema SQL, and tests for migration ordering and core table coverage. Runtime application of migrations is still pending in the repository layer. |
| SEC-004 Domain models | TESTED | Added canonical world, branch, club, player, season, fixture, match, and event models with tests for position parsing, fixture validation, and display name rules. |
| SEC-005 Repository layer | TESTED | Added repository contracts plus a SQLite SQL store with migration execution, create/list methods for core entities, and store-level tests using fakes to verify mapping and migration flow. |
| SEC-007 TUI app shell and navigation | TESTED | Replaced the placeholder shell with a Bubble Tea app using Lip Gloss and Bubbles-based key bindings, with tests covering navigation and rendered shell regions. |
| SEC-020 Lineup selection engine | TESTED | Added deterministic formation templates and lineup assignment with locked-player preassignment, goalkeeper constraints, and tests for slot coverage, formation support, and secondary-position use. |
| SEC-021 Match engine tick loop | TESTED | Added a deterministic match loop skeleton with team plan derivation, possession flow, chance/goal event generation, and repeatability tests across identical inputs. |
| SEC-022 Match action resolution | TESTED | Replaced aggregate chance events with deterministic action phases covering build-up, penetration, shot attempts, saves, blocks, goals, and turnovers, with ordering and event-type tests. |
| SEC-023 Fatigue system | TESTED | Added deterministic per-player fatigue accumulation by slot and possession load, exposed average fatigue in match summaries, and tested that fatigue builds over time and reduces effective team strength. |
| SEC-024 Injury engine | TESTED | Added deterministic injury triggering from fatigue/load, injury event emission, stable injury summaries, and tests for deterministic injury records and high-fatigue trigger conditions. |
| SEC-025 Cards and suspension system | TESTED | Added deterministic yellow/red card generation, tracked red-card suspensions in match summaries, and tested deterministic card output plus high-fatigue trigger conditions. |
| SEC-026 Set piece engine | TESTED | Added deterministic corner, free-kick, and penalty branches in action resolution, including set-piece goals and clearances, with event-type and trigger-path tests. |
| SEC-027 Match persistence and stats aggregation | TESTED | Added aggregate match stats to summaries and a simulation service helper that persists match rows plus event logs through the repository contracts, with service tests for event backfilling and error propagation. |
| SEC-011 Match Lab screen | TESTED | Added interactive Match Lab screen with two-pane layout (setup + results), demo squads, formation cycling, seed control, simulation trigger, event log scrolling, and 10 passing tests. Screen interface + adapter pattern added for future screens. |
| SEC-012 Match result and event log screens | TESTED | Extended Match Lab with two full-screen drill-down views: result detail (v key — expanded stats, fatigue, injuries, cards, suspensions, scrollable) and event log (e key — full scrollable event list with f-key filter cycling across All/Goals/Cards/Injuries/Set Pieces). b key returns to split view from either sub-view. 13 new passing tests. |
| SEC-028 Fixture generation | TESTED | Implemented Berger-table round-robin fixture generation for N clubs (even/odd), home-and-away return legs, weekly date spacing, and 8 passing tests covering pair coverage, matchday range, self-fixture exclusion, and schedule correctness. |
| SEC-029 League table engine | TESTED | Implemented 3-1-0 points standings computation from match results, with tie-breaking by GD/GF/name, and 6 tests covering win/draw/loss, unknown fixtures, and sort correctness. |
| SEC-030 Season progression loop | TESTED | Implemented deterministic full-season RunSeason loop that generates lineups per fixture and produces SeasonResult with results + table; tests cover two-club season, determinism, and missing squad skip. |
| SEC-031 Recovery and between-match updates | TESTED | Implemented fatigue recovery (configurable per-day rate), injury recovery days by severity (minor/moderate/major), and FilterActiveInjuries; 3 tests covering fatigue floor, injury filtering at 7/21/42 days. |
| SEC-013 Season Lab screen | TESTED | Added Season Lab TUI screen with setup/standings/results views for a 6-club demo league, seed control, deterministic simulation, and 11 passing tests. |
| SEC-008 Dashboard screen | TESTED | Added Dashboard TUI screen with world summary, season progress bar, and scrollable layout; 5 tests. |
| SEC-009 Squad screen | TESTED | Added Squad TUI screen with sortable player roster (OVR/POS/Name), cursor navigation, and 6 tests. |
| SEC-010 Player detail screen | TESTED | Added Player Detail TUI screen with attribute bars, player browse (h/l), scroll, and 4 tests. |
| SEC-014 Club screen | TESTED | Added Club TUI screen with club profile, squad overview, recent form (W/D/L coloured), and 3 tests. |
| SEC-032 Player development engine | TESTED | Implemented age-based attribute growth (youth→peak), peak hold (26–29), decline (30+), with floor clamping and 5 tests covering each phase and the 40-floor invariant. |
| SEC-033 Retirement and aging system | TESTED | Implemented AgeAtDate, ShouldRetire (≥35), ProcessSquadAging with retirement filtering; 4 tests including retirement segregation and DOB calculation. |
| SEC-034 Youth intake / regen system | TESTED | Implemented deterministic GenerateYouthIntake with configurable count, potential range, and position pool; 3 tests covering count, determinism, and potential bounds. |
| SEC-035 Club finance engine | TESTED | Implemented annual revenue calculation (ticket/TV/sponsor scaled by league position), closing balance, and 40%-of-surplus transfer budget; 5 tests. |
| SEC-036 Transfer valuation and decision engine | TESTED | Implemented MarketValue (OVR²-based with age multiplier and potential bonus) and EvaluateTransfer (buy/pass AI decision); 6 tests. |
| SEC-037 Contract logic | TESTED | Implemented Contract, ContractFor, RenewalOffer, AnnualWageBill, ExpiringContracts, and status classification (secure/monitor/expiring); 7 tests. |
| SEC-015 Transfers screen | TESTED | Added Transfers TUI screen with Shortlist and Contracts tabs, cursor navigation, and 5 tests. |
| SEC-017 World screen | TESTED | Added World TUI screen with world summary and competition standings overview; 3 tests. |
| SEC-016 Scenarios / branch management screen | TESTED | Added Scenarios TUI screen with branch list, active branch indicator, enter-to-activate, and 4 tests. |
| SEC-038 Ingestion raw staging pipeline | TESTED | Implemented InMemoryStagingStore with StageRecord, ListStaged (with source/type filters and processed exclusion), MarkProcessed; 3 tests. |
| SEC-039 Source adapters | TESTED | Implemented SourceAdapter interface and StaticAdapter; tested via StaticAdapter.Fetch staging 2 records. |
| SEC-040 Normalization pipeline | TESTED | Implemented Normalize func with NormalizerFunc type and error collection; 1 test. |
| SEC-041 Entity resolution | TESTED | Implemented EntityResolver with name-canonicalization index, Resolve returning match score and new-entity flag; 2 tests. |
| SEC-042 Validation and publish pipeline | TESTED | Implemented ValidatePlayerRecord and ValidateClubRecord with required-field checks; 3 tests. |
| SEC-018 Data / import screen | TESTED | Data/Import reads `ingest_state.json`; keys `r` reload, `f` demo fetch, `s` FBref scrape (requires `fbref_url.txt` with one URL line); `app.Config.StateDir` + `ResolveStateDir()`; tests use `NewWithStateDir`. |
| SEC-043 Python calibration bridge | TESTED | Implemented file-based Bridge with WriteRequest, ReadResult, PendingRequests using JSON I/O; 4 tests including write/read roundtrip and pending detection. |
| SEC-044 Export tools | TESTED | Implemented ExportPlayersCSV, ExportStandingsCSV, ExportMatchResultsJSON; 3 tests. |
| SEC-045 Test harness and realism regression suite | TESTED | Implemented TestRealismRegression covering match count, table structure, GD consistency, points accounting (3pts/win + 1pt/draw), and full determinism across two identical runs. |
| SEC-046 Documentation and handoff hygiene | DONE | Updated TASKS.md (all tasks marked complete), IMPLEMENTATION_TRACKER.md (full ledger), and this snapshot. |
| SEC-047 FBref HTML scraper adapter | TESTED | `internal/ingestion/fbref` parses `table.stats_table` tbody rows with `th[data-stat=player]` + `/players/` links; `NewFbrefPlayerTableAdapter` HTTP GET with timeout + UA; respects staging cap via `RunAdapters`; tests use `httptest` + `testdata/minimal_table.html`. |
| SEC-048 FBref normalizer | TESTED | `NormalizeFbrefPlayerRecord` + `mapFBrefPosition` map raw JSON to `NormalizedRecord` and `ValidatePlayerRecord`-ready attrs; optional `overall_approx` heuristic from goals/assists/minutes keys. |
| SEC-049 CLI scrape command | TESTED | `simcli scrape [--url U] [--state-dir D]` or URL from `fbref_url.txt`; updates same ingest ledger as `fetch`. |
| SEC-050 Data/Import scrape trigger | TESTED | Key `s` runs FBref scrape when `fbref_url.txt` exists in state dir; `f`/`r` unchanged; help text updated. |

## Additional progress

- SQLite world bootstrap | TESTED | `cmd/simtui` now opens `football.db` under the resolved state dir via `modernc.org/sqlite`, applies migrations, seeds a default world/branch/clubs/players when empty, and passes the active repository + IDs through `app.Config`. Dashboard, World, Club, Squad, and Player Detail now read from the seeded SQLite state with demo fallback when repository data is unavailable. Added app bootstrap tests plus repository timestamp-scan fixes for real SQLite rows. |
- Match/season persistence wiring | TESTED | Match Lab and Season Lab now load the active branch clubs/squads from SQLite when at least two clubs have 11 players, falling back to demo data otherwise. Running either screen persists seasons, fixtures, matches, and event logs through the repository layer. Bootstrap now also repairs seeded squads up to 11 players so the storage-backed match flow remains runnable across existing local databases. |
- TUI scraper integration | TESTED | Data / Import now exposes an in-screen FBref URL editor using Bubble Tea text input. `u` enters edit mode, `enter` saves the URL to `fbref_url.txt`, and `s` scrapes the currently saved URL directly from the TUI without requiring a separate CLI step. Added screen coverage for URL save flow. |
- History/replay and SQLite ingest staging | TESTED | Added repository reads for seasons and matches by season, Season Lab saved-season reload on `h`, Match Lab saved-match replay on `p`, home/away club cycling on `[ ]` and `{ }`, and a SQLite-backed staging store over `raw_sources` / `raw_payloads` with `processed_at` tracking via migration `0002_raw_payload_processing.sql`. `RunDemoFetch` and FBref scrape now stage into SQLite instead of in-memory storage. |
- Shared branch state and replay branching | TESTED | Added `app.RuntimeState` so screens can resolve the current world/branch/club through shared mutable app state. Scenarios now loads repository branches and updates the shared active branch on `enter`. Data / Import previews staged raw payload rows for the selected source. Match Lab can create a new child branch from the selected saved match on `c`, and subsequent screens resolve branch-aware state through `Config.Current*()` helpers. |
- Normalized preview and season replay branching | TESTED | Data / Import now renders normalized FBref preview rows from staged SQLite payloads using `NormalizeFbrefPlayerRecord` plus validation hints instead of raw-only browsing. Season Lab now supports creating a replay branch from the selected saved season on `c`. Match replay reconstruction also restores cards, injuries, and red-card suspensions from persisted event logs so saved match detail views are closer to live simulation output. |
- Publish-candidate preview | TESTED | Data / Import now derives a lightweight publish preview from staged rows: normalized entity names, validation state, and provisional resolved IDs via `EntityResolver`. This gives the TUI a basic published-entity candidate browser even before a full publish workflow exists. |
- SQLite publish action and browser | TESTED | Added migration `0003_published_entities.sql`, a SQLite-backed published-entity store, `PublishSourceRecords`, and persisted-row updates. Data / Import now publishes the selected source on `p`, marks staged rows processed, renders persisted published entities from the SQLite publish store alongside raw/normalized/publish-candidate previews, and supports in-screen detail review plus `name` / `resolved_id` edits for published rows. |
- Published entity review/edit controls | TESTED | Data / Import now allows keyboard review of persisted published entities with a dedicated detail pane, `tab` focus switching, and inline `name` / `resolved_id` edits backed by `UpdatePublished` in the SQLite staging store. Added store and screen coverage for persisted-row updates. |

## Risks and caveats

- The match engine (Phase B) uses deterministic seed-based scoring without calibrated probability distributions. Goals-per-match (~15) exceeds real-world rates (~2.5). The Python calibration bridge (SEC-043) provides the interface for tightening these rates via the mlbridge package.
- Saved history can now be reloaded, but replay currently reconstructs a lightweight summary from persisted match rows plus event logs rather than a richer fully persisted stats model.
- Branch activation is now wired into shared runtime state, and replay-branch creation exists in Match Lab and Season Lab, but broader timeline editing/checkpoint workflows are still missing.
- Fetch metadata is persisted in `ingest_state.json` (per-source caps default 2500/run, editable in the file). `simcli scrape` performs real HTTP to user-supplied FBref URLs; obey site terms, robots.txt, and rate limits. The TUI now shows raw, normalized, publish-candidate, and persisted published-entity views, and supports lightweight review/edit controls for published rows, but there is still no approval workflow or import into branch/world state from those published entities.
