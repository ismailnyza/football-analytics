# Skill: Data Ingestion Work

## Use this skill when
- building source adapters
- parsing raw payloads
- normalizing entities
- implementing validation and publish flows

## Goals
- raw data retention
- explicit normalization rules
- explicit confidence tracking
- safe publish flow

## Rules
- preserve raw source payloads in staging tables
- make source-specific logic local to adapter packages
- normalize into canonical models
- never silently accept low-confidence merges
- publish only after validation

## Missing data policy
If data is missing:
1. use source fallback
2. use explicit inference
3. use low-confidence default only if necessary
4. record confidence metadata

## FIFA / FM note
FIFA / EA ratings and Football Manager data are optional enrichment only.
They are not mandatory and not the sole source of truth.

## Deliverables
- adapter implementation
- normalizer changes
- validation logic
- tracker update
- passing verification before commit/push
