# Skill: Core Simulation Work

## Use this skill when
- implementing match engine logic
- implementing lineup logic
- implementing fatigue, injuries, or set pieces
- changing deterministic simulation rules

## Goals
- preserve deterministic behavior
- keep formulas explicit
- keep match state debuggable
- avoid hidden randomness

## Rules
- always work from canonical domain models
- use seed-driven RNG only
- document formula changes in tracker notes
- prefer discrete stable logic over overfitted complexity

## Required checks
- same seed + same inputs => same outputs
- no UI dependency in simulation packages
- event generation aligns with persisted stats

## Deliverables
- engine code
- tests for changed formulas
- `TASKS.md` update
- `docs/IMPLEMENTATION_TRACKER.md` update
- passing verification before commit/push
