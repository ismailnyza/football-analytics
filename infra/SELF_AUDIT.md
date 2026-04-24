# Self-Audit Log

## Audit 2026-04-25 — Iteration 0
### Prompt review
- The repository now has a local master-prompt copy, but no validated evidence claims have been copied into code or docs.
- The first loop should optimize for truthful baseline creation, not ambitious feature sprawl.

### Skill review
- Added repo-local skills for agent infrastructure and evidence grading.
- No deprecated repo-local skills yet.

### Memory-style ledger review
- Evidence, disproven, and factor graph ledgers were initialized empty on purpose.
- No claim should move into the engine until the first backtest exists.

### Workflow review
- Current bottleneck: zero empirical baseline.
- Next workflow improvement: create a repeatable baseline data contract and backtest command.

## Audit 2026-04-25 — Iteration 1
### Prompt review
- The empirical loop now has a real checkpoint: an EPL historical-team-strength baseline measured on a named holdout season.
- The next prompt pressure should focus on residual-specific error reduction rather than broad new feature sprawl.

### Skill review
- Added `.skills/backtest-runner` because the repo now has a repeatable validation command and artifact path.
- Existing repo-local skills remained useful and did not require deprecation.

### Memory-style ledger review
- The evidence matrix now contains the first measured entry.
- No disproven factors were added yet, but a concrete model failure is known: zero predicted draws.

### Workflow review
- Improvement: the repo can now go from checked-in CSVs to a deterministic validation artifact with one command.
- New bottleneck: the draw mechanism is underfit, causing class-distribution mismatch on holdout evaluation.

## Audit 2026-04-25 — Iteration 2
### Prompt review
- Residual-driven work paid off immediately: targeting the draw miss uncovered a stronger candidate family than the original Davidson backtest.
- The prompt should continue to favor targeted error reduction over generic feature growth.

### Skill review
- The new Python search script is a useful offline calibration workflow and supports the existing backtest-runner skill rather than requiring a new deprecation cycle.

### Memory-style ledger review
- The evidence matrix now distinguishes committed baseline evidence from promising but not-yet-integrated offline search evidence.

### Workflow review
- Improvement: residual analysis now has a reproducible search script and artifact path.
- New bottleneck: the stronger draw-decay family exists only as an offline search result until it is ported into the main Go harness.

## Audit 2026-04-25 — Iteration 3
### Prompt review
- The draw-decay baseline is now integrated into the main Go backtest path (SEC-009 TESTED).
- A naive-frequency floor benchmark is implemented (SEC-010 TESTED), giving a measurable sanity-check baseline at 40.79%.
- Player-level domain models exist as a foundation (SEC-012 TESTED), marking the transition from pure team-level to atomic match modelling.
- The prompt should continue favoring evidence-driven expansion rather than speculative features.

### Skill review
- Existing skills remain useful; no deprecations needed.
- The backtest-runner skill now covers both baseline-backtest and naive-frequency commands.

### Memory-style ledger review
- Evidence matrix updated to reflect the integrated draw-decay baseline at Grade B.
- Disproven cemetery now records the Davidson fixed-draw factor failure.
- Factor graph shows active nodes and planned expansion.

### Workflow review
- Improvement: the repo went from zero-draw baseline to non-zero-draw with a measured floor benchmark, plus player-level domain model foundation — all within iteration 3.
- New bottleneck: draw under-prediction remains severe (3.42% vs actual 24.47%).
- Next bottleneck: no cross-league validation, no player-level data ingestion.
