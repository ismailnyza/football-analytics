# Football Simulation Engine Architecture

## Purpose

This document defines the technical design boundaries and primary module contracts.

---

## System overview

The application is a terminal-first football simulation engine with a Go runtime and optional Python offline calibration support.

### Runtime ownership

#### Go owns
- TUI
- simulation engine
- persistence
- branch management
- season progression
- finance logic
- transfer logic
- export pipeline

#### Python owns
- coefficient fitting
- model calibration
- xG lookup generation
- offline data science experiments
- optional support scripts

---

## High-level modules

### `internal/domain`
Pure domain models and shared business entities.
No storage or UI dependencies.

### `internal/engine`
Simulation logic.
Submodules:
- `match`
- `season`
- `injury`
- `development`
- `transfer`
- `finance`

### `internal/storage`
Persistence layer.
Submodules:
- `sqlite`
- `postgres`
- `migrations`

### `internal/tui`
All terminal UI logic.
Submodules:
- `screens`
- `components`
- `layout`
- `keymap`

### `internal/services`
Application orchestration layer between storage, engine, and UI.

### `internal/ingestion`
Raw source fetching, parsing, normalization, validation.

### `internal/mlbridge`
Loads offline-generated model artifacts into Go runtime.

---

## Match engine model

A match is simulated over fixed ticks.

Recommended V1 model:
- 900 ticks
- 6 seconds per tick
- discrete pitch zones
- contextual actions based on phase and zone

### Match flow
1. initialize lineups
2. initialize effective attributes
3. iterate ticks
4. resolve phase and action
5. persist events and stats
6. finalize result

---

## Branching model

Simulation branches must behave like alternate timelines.

Rules:
- root world remains canonical
- edited scenarios create child branches
- replayed matches must never overwrite parent branch history
- overrides must be explicit and queryable

---

## TUI layout contract

Global layout:
- top bar
- left navigation
- main content view
- bottom status bar

Primary screens:
- Dashboard
- Match Lab
- Season Lab
- Club
- Squad
- Player Detail
- Transfers
- World
- Scenarios
- Data / Import
- Settings

---

## Persistence strategy

Default database is SQLite.

Why:
- zero setup
- local-first
- ideal for a TUI tool
- easy snapshotting and export

Postgres may be added later behind the same repository interfaces.

---

## Data pipeline strategy

### Ingestion stages
1. raw fetch
2. source parsing
3. normalization
4. inference
5. validation
6. publish season snapshot

### Important rule
Raw source data must be retained in staging tables.
Do not make normalization depend on transient memory-only transformations.

---

## Runtime config and model artifacts

Offline-generated artifacts may include:
- `xg_lookup.json`
- `injury_coefficients.json`
- `development_curve.json`
- `transfer_weights.json`

These are loaded by Go at runtime.
The app must still run if some optional artifacts are absent.

---

## Testing strategy

### Unit tests
- pass resolution
- shot resolution
- tackle/foul logic
- injury probability
- development deltas

### Integration tests
- one match end-to-end
- one matchweek
- one season
- replay branch creation

### Realism regression
Track broad ranges for:
- goals per match
- home win rate
- draw rate
- injury rate
- top scorer range
- card frequency

---

## Architecture rule

If a change crosses domain, engine, storage, TUI, and Python all at once, it is probably too big.
Split it.
