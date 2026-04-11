# Skill: Python Calibration Work

## Use this skill when
- building coefficient generation scripts
- fitting realism parameters
- generating lookup tables
- experimenting with calibration data

## Goals
- produce offline artifacts that Go can load
- improve realism without introducing runtime coupling
- keep file formats stable and documented

## Rules
- Python outputs must be versioned artifacts
- Go runtime must not require live Python execution for normal usage
- artifact schema must be documented
- calibration scripts must be reproducible

## Common outputs
- xg_lookup.json
- injury_coefficients.json
- development_curve.json
- transfer_weights.json

## Deliverables
- script or notebook
- artifact schema note
- generated output path
- `TASKS.md` update
- `docs/IMPLEMENTATION_TRACKER.md` update
- passing verification before commit/push
