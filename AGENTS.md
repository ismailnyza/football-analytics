# Agent Working Agreement

Read these files first:
1. `SKILL.md`
2. `TASKS.md`
3. `ARCHITECTURE.md`
4. `docs/IMPLEMENTATION_TRACKER.md`
5. relevant `.skills/*/SKILL.md`

Working style
- make focused changes
- keep the repository buildable
- run relevant tests after every change
- update `TASKS.md`, `PLANS.md`, and `docs/IMPLEMENTATION_TRACKER.md` honestly
- keep the Go runtime deterministic
- keep Python offline-only for calibration and research
- prefer explicit evidence tracking over undocumented intuition
- commit only passing work
- push passing commits when remote access is available

Current mission
- bootstrap a self-improving football simulation research repository from scratch
- build the agent infrastructure that can revise prompts, skills, and evidence ledgers
- start with a minimal executable baseline and a truthful research backlog
- prioritize a historical-team-strength baseline before richer match-engine factors

Required output after work
```text
Handoff Summary
- Sections touched:
- Files changed:
- Build status:
- Test status:
- Remaining work:
- Risks / caveats:
```

Do not
- fabricate evidence grades or validation results
- claim tests that were not run
- add Python into the interactive runtime loop
- silently skip tracker updates
