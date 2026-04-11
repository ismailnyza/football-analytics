# Football Simulation Engine Skill

## Purpose

This file is the primary operating contract for coding agents working on this repository.

Agents must read this file before making changes.

---

## Project identity

This project is a **local-first terminal football simulation engine**.

It is built for:
- deterministic match simulation
- season progression
- multi-year club simulation
- lineup and scenario experimentation
- branchable alternate worlds

It is **not**:
- a web app
- a mobile app
- a 3D match renderer
- a live-service multiplayer game

---

## Approved stack

### Core
- Go
- Bubble Tea
- Bubbles
- Lip Gloss
- SQLite
- Cobra

### Optional / advanced
- Postgres
- sqlc
- Python offline calibration tools

### Python rule
Python may be used for:
- xG lookup generation
- injury coefficient fitting
- transfer valuation calibration
- realism tuning
- notebooks
- support scripts

Python must not be used for:
- TUI hot path
- match tick loop runtime dependency
- required synchronous UI interactions

---

## Project priorities

### Priority 1
- correctness
- determinism
- maintainability
- terminal usability

### Priority 2
- realism
- extensibility
- performance

### Priority 3
- polish
- secondary convenience features

---

## Working rules

### Rule 1: Keep boundaries clean
Do not mix:
- TUI rendering with engine logic
- storage code with domain rules
- Python workflows with Go runtime flow

### Rule 2: Preserve determinism
Core simulation must be reproducible with the same seed and same inputs.

### Rule 3: Keep the app runnable
Avoid giant destabilizing rewrites.
Small, coherent changes are preferred.

### Rule 4: Use tracker truthfully
When a task is changed, update:
- `TASKS.md`
- `docs/IMPLEMENTATION_TRACKER.md`

Never claim work is tested if it was not tested.
Never finish a section without updating both files to reflect the real status.

### Rule 5: Respect the TUI
This is keyboard-first software.
Design for speed, clarity, and focus.
Do not create UI patterns that assume mouse-first interaction.

### Rule 6: Verify before publish
Every code change must be validated with relevant automated checks before it is considered done.
If a check fails, fix the issue and rerun the checks.
Only commit after the checks pass.
If a remote branch exists, push the passing commit before handing off.

### Rule 7: Reuse approved permissions
When command permissions or prefix approvals have already been granted in this workspace, agents should reuse them instead of asking again for the same capability.
Treat existing approvals as durable for later agent work on the same repository unless the action is materially broader or more destructive.

---

## Required reading order for agents

Before making changes, read in this order:

1. `SKILL.md`
2. `TASKS.md`
3. `ARCHITECTURE.md`
4. `docs/IMPLEMENTATION_TRACKER.md`
5. relevant `.skills/*/SKILL.md`

---

## Expected repository structure

```text
cmd/
  simtui/
  simcli/

internal/
  app/
  tui/
    screens/
    components/
    layout/
    keymap/
  domain/
  engine/
    match/
    season/
    injury/
    development/
    transfer/
    finance/
  storage/
    sqlite/
    postgres/
    migrations/
  services/
    simulation/
    replay/
    scouting/
    importers/
  ingestion/
    fbref/
    understat/
    transfermarkt/
    whoscored/
    normalizer/
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

docs/
.skills/
```

---

## Decision rules

When uncertain:

### If it affects simulation correctness
Prefer clearer deterministic logic over clever complexity.

### If it affects TUI usability
Prefer fewer screens with clearer panes over fragmented screen sprawl.

### If data is missing
Use:
1. explicit fallback
2. explicit inference
3. confidence marking

Never silently invent hidden data paths.

---

## TUI design rules

The TUI must have:
- app shell
- top bar
- side navigation
- main content area
- bottom help/status bar

Preferred screens:
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

Keybinding defaults:
- `j/k` or arrows for movement
- `tab` for pane switching
- `enter` to select
- `e` to edit
- `s` to simulate
- `b` to create/view branch
- `f` to filter
- `/` to search
- `?` for help
- `esc` back

---

## Definition of done

A unit of work is only done when:
- implementation exists
- build still works
- acceptance criteria are satisfied
- relevant tests or verification commands were run and passed
- failing checks were fixed before completion
- tracking files were updated
- `TASKS.md` and `docs/IMPLEMENTATION_TRACKER.md` were both updated
- the passing change was committed
- the passing commit was pushed when push access is available
- already-approved command permissions were reused when applicable

---

## Deliverable format expected from agents

When completing work, provide a short handoff summary:

```text
Handoff Summary
- Sections touched:
- Files changed:
- Build status:
- Test status:
- Remaining work:
- Risks / caveats:
```
