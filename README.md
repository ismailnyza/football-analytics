# Self-Improving Football Simulation Research Agent

This repository is the fresh bootstrap for a football simulation research system and the agent infrastructure that improves it.

Mission
- build a football simulation engine from empirical evidence
- keep the runtime deterministic and local-first
- improve the agent infrastructure itself: prompts, skills, memory hygiene, and workflow

Current state
- iteration 0 bootstrap complete
- repository scaffolded for engine, calibration, evidence tracking, and self-audit
- first executable artifact is a tiny `simcli` status command plus evidence primitives
- first research target is a historical team-strength baseline for match result prediction

Core directories
- `engine/` — Go packages for evidence tracking and simulation code
- `calibration/` — offline experimentation, backtests, and factor testing
- `docs/` — theory, factors, validation, and research notes
- `infra/` — prompt/skill/memory audit ledgers
- `data/` — checked-in datasets and provenance notes
- `video/` — video verification notes and future tooling
- `.skills/` — repo-local workflows for future agents

Quick start
```bash
make test
make build
go run ./cmd/simcli status
```
