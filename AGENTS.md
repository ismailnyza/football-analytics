# Agent Working Agreement

Read these files first:
1. `SKILL.md`
2. `TASKS.md`
3. `ARCHITECTURE.md`
4. `docs/IMPLEMENTATION_TRACKER.md`
5. relevant `.skills/*/SKILL.md`

## Working style
- make focused changes
- keep build green
- run relevant tests after every change
- if tests fail, fix them before moving on
- update trackers honestly
- respect Go/Python boundary
- respect TUI-first design
- commit only passing work
- push passing commits when remote access is available

## Required output after work

```text
Handoff Summary
- Sections touched:
- Files changed:
- Build status:
- Test status:
- Remaining work:
- Risks / caveats:
```

## Do not
- silently change architecture
- silently change simulation semantics
- claim tests that were not run
- add Python runtime dependency into the TUI loop
- commit failing work as complete
