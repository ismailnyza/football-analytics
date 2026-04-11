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
- [ ] SEC-005 Repository layer
- [ ] SEC-007 TUI app shell and navigation

---

## Phase B: Core Match Simulation

- [ ] SEC-020 Lineup selection engine
- [ ] SEC-021 Match engine tick loop
- [ ] SEC-022 Match action resolution
- [ ] SEC-023 Fatigue system
- [ ] SEC-024 Injury engine
- [ ] SEC-025 Cards and suspension system
- [ ] SEC-026 Set piece engine
- [ ] SEC-027 Match persistence and stats aggregation
- [ ] SEC-011 Match Lab screen
- [ ] SEC-012 Match result and event log screens

---

## Phase C: Season Simulation

- [ ] SEC-028 Fixture generation
- [ ] SEC-029 League table engine
- [ ] SEC-030 Season progression loop
- [ ] SEC-031 Recovery and between-match updates
- [ ] SEC-013 Season Lab screen
- [ ] SEC-008 Dashboard screen
- [ ] SEC-009 Squad screen
- [ ] SEC-010 Player detail screen
- [ ] SEC-014 Club screen

---

## Phase D: Long-Term World Simulation

- [ ] SEC-032 Player development engine
- [ ] SEC-033 Retirement and aging system
- [ ] SEC-034 Youth intake / regen system
- [ ] SEC-035 Club finance engine
- [ ] SEC-036 Transfer valuation and decision engine
- [ ] SEC-037 Contract logic
- [ ] SEC-015 Transfers screen
- [ ] SEC-017 World screen
- [ ] SEC-016 Scenarios / branch management screen

---

## Phase E: Data and Calibration Layer

- [ ] SEC-038 Ingestion raw staging pipeline
- [ ] SEC-039 Source adapters
- [ ] SEC-040 Normalization pipeline
- [ ] SEC-041 Entity resolution
- [ ] SEC-042 Validation and publish pipeline
- [ ] SEC-018 Data / import screen
- [ ] SEC-043 Python calibration bridge
- [ ] SEC-044 Export tools
- [ ] SEC-045 Test harness and realism regression suite
- [ ] SEC-046 Documentation and handoff hygiene

---

## Immediate next task

`SEC-005 Repository layer`

---

## Notes

- `SEC-001` and `SEC-002` are complete only as a minimal bootstrap.
- `SEC-003` currently covers embedded SQL migrations and schema definition.
- `SEC-004` currently covers canonical entity structs and basic invariants.
- Runtime SQLite execution wiring is still pending under the repository layer.
- Always update `docs/IMPLEMENTATION_TRACKER.md` when task status changes.
