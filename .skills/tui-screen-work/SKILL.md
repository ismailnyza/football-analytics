# Skill: TUI Screen Work

## Use this skill when
- building screens
- building components
- defining keybindings
- adjusting layouts
- improving navigation flows

## Goals
- keyboard-first usability
- clean information density
- consistent layout patterns
- low-friction simulation workflow

## Rules
- keep shell consistent across screens
- avoid visual clutter
- prefer split panes, tabs, and tables
- do not put business logic in render layer
- use status bar for key hints

## Screen expectations
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

## Required checks
- navigation works without mouse
- selected item state is visible
- empty/loading/error states exist
- screen does not require future web UI assumptions

## Deliverables
- screen model
- update function
- view function
- reusable components if repeated
- `TASKS.md` update
- `docs/IMPLEMENTATION_TRACKER.md` update
- passing verification before commit/push
