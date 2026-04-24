# Implementation Tracker

## Snapshot
- Branch state: dev
- Runtime state: Fresh bootstrap with minimal Go CLI and evidence primitives
- Last verified commands:
  - `go test ./...`
  - `go build ./...`
  - `go run ./cmd/simcli status`

## Task ledger
| Task | Status | Notes |
| --- | --- | --- |
| SEC-000 Preserve git history and wipe the legacy working tree | TESTED | Removed previous working-tree contents while keeping `.git`, then rebuilt the repo from scratch. |
| SEC-001 Reinitialize repository structure | TESTED | Added clean top-level directories and foundational docs for engine, calibration, data, docs, infra, and video workflows. |
| SEC-002 Create infra ledgers | TESTED | Added `SKILL_REGISTRY.md`, `PROMPT_CHANGELOG.md`, `SELF_AUDIT.md`, `DISPROVEN_CEMETERY.md`, `EVIDENCE_MATRIX.md`, and `FACTOR_GRAPH.md`. |
| SEC-003 Minimal Go baseline | TESTED | Added evidence-grade primitives, factor validation rules, tests, and a `simcli status` command. |
| SEC-004 Initial research notes and repo-local skills | TESTED | Added iteration-0 audit/research notes and repo-local skills for agent infrastructure and evidence grading. |

## Risks and caveats
- No football prediction model exists yet; only the bootstrap infrastructure is present.
- No empirical factor has been validated yet, so `infra/EVIDENCE_MATRIX.md` intentionally contains only planned work.
- The repo is honest but very early: the next real milestone is the historical-team-strength baseline.
