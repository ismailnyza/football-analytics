# Implementation Tracker

This file is the single truth ledger for implementation, review, and test status.

## Snapshot

- Branch state: reset to clean bootstrap on 2026-04-11
- Runtime state: Go-first scaffold only
- Legacy Python runtime: removed from this branch
- Last verified commands:
  - `go build ./...`
  - `go test ./...`
  - `go vet ./...`

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

## Risks and caveats

- The branch is intentionally reset; previous Python implementation files were removed rather than migrated.
- The project now depends on Bubble Tea, Bubbles, and Lip Gloss for the TUI shell.
- SQLite repository logic exists, but no external SQLite driver dependency has been added yet for live integration tests.
- The next task should layer a deterministic injury engine over the current fatigue-aware action loop.
