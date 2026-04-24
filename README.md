# Self-Improving Football Simulation Research Agent

This repository is the fresh bootstrap for a football simulation research system and the agent infrastructure that improves it.

Mission
- build a football simulation engine from empirical evidence
- keep the runtime deterministic and local-first
- improve the agent infrastructure itself: prompts, skills, memory hygiene, and workflow

Current state
- iteration 3 complete: draw-decay baseline integrated into main Go harness
- committed holdout accuracy: 53.42% W/D/L on 2024/25 EPL (with non-zero draw predictions)
- best offline candidate: 54.74% (weaker validation support)
- player-level domain models defined as foundation for atomic match modelling
- next: naive-frequency benchmark (SEC-010), cross-league validation (SEC-011)

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
