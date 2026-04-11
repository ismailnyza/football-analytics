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
| SEC-005 Repository layer | NOT_STARTED | No storage interfaces or adapters yet. |
| SEC-007 TUI app shell and navigation | NOT_STARTED | Current `simtui` is a placeholder text shell, not a Bubble Tea app. |

## Risks and caveats

- The branch is intentionally reset; previous Python implementation files were removed rather than migrated.
- The current bootstrap avoids external dependencies so the repo stays buildable in a restricted environment.
- SQLite schema exists, but no driver-backed migrator has been wired yet.
- The next task should introduce repository interfaces and a SQLite-backed persistence layer on top of the current schema and domain models.
