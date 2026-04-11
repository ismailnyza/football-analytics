# Skill: Realism Regression Work

## Use this skill when
- building regression tests
- evaluating simulation output ranges
- tuning match or season realism

## Goals
- catch realism drift
- provide broad statistical guardrails
- avoid overfitting to one league or one season

## Metrics to watch
- goals per match
- home win rate
- draw rate
- red card rate
- injury rate
- possession spread
- top scorer range
- league points distribution

## Rules
- use broad bands, not brittle exact values
- compare aggregate outputs over enough simulations
- document thresholds in tracker or docs

## Deliverables
- regression suite
- metric thresholds
- notes on failed realism checks
- tracker update
- passing verification before commit/push
