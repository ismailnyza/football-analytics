# Football Simulation Engine

Local-first terminal football simulation engine.

## Current state

This branch has been reset to a clean bootstrap aligned with the Go-first architecture.

Implemented so far:
- repository scaffolding for the target module layout
- minimal Go module and runnable entrypoints
- project tracker and task queue reset to the current truth

Not implemented yet:
- SQLite schema and migrations
- domain models
- repository layer
- Bubble Tea TUI shell
- match and season simulation logic

## Repository guide

- `SKILL.md`: primary project rules
- `TASKS.md`: ordered implementation queue
- `ARCHITECTURE.md`: technical boundaries
- `AGENTS.md`: working agreement
- `docs/IMPLEMENTATION_TRACKER.md`: implementation ledger
- `.skills/`: task-specific workflows

## Build

```bash
make build
make test
make run-cli
make run-tui
```

## Layout

```text
cmd/
  simtui/
  simcli/

internal/
  app/
  domain/
  engine/
  storage/
  tui/
  services/
  ingestion/
  export/
  mlbridge/

python/
  calibration/
  models/
  pipelines/
  notebooks/

data/
  seeds/
  configs/
  models/
```

