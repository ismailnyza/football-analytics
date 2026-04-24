# Football Simulation Research Skill

## Purpose
This is the operating contract for coding and research agents working in this repository.

## Project identity
This repository is a self-improving football simulation research system.

It exists to:
- build a deterministic football simulation engine in Go
- evaluate candidate factors with explicit evidence grading
- maintain agent infrastructure that can audit prompts, skills, and memory assumptions

It is not yet:
- a finished simulator
- a production scouting platform
- a validated prediction system

## Approved stack
Core runtime
- Go

Offline research and calibration
- Python
- notebooks or scripts under `calibration/`

## Working rules
1. No false certainty
   - hypotheses may be documented, but only validated factors may influence the engine
2. Determinism first
   - same seed and same inputs must yield the same runtime outputs
3. Tracker truthfulness
   - when work changes status, update `TASKS.md`, `PLANS.md`, and `docs/IMPLEMENTATION_TRACKER.md`
4. Infrastructure matters
   - if prompt, skill, or workflow assumptions are revised, log them under `infra/`
5. Small coherent changes
   - avoid giant speculative rewrites without a measured baseline

## Definition of done
A unit of work is done only when:
- the implementation or document exists
- relevant checks were run and passed
- tracker files were updated truthfully
- the result is committed
- the passing commit is pushed when possible
