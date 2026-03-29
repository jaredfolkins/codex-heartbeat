# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Seed fresh autoresearch workspaces with generic guardrail sections in scaffolded `PLANNING.md` so safe parity-gap work starts with blocked/non-goals and acceptance criteria by default.
- Primary evaluator: `go test ./cmd/codex-heartbeat -run 'EnsureAutoresearchWorkspaceScaffoldsWorkspace|EnsureAutoresearchWorkspaceSeedsPlanningTaskList|EnsureAutoresearchWorkspaceSeedsPlanningGuardrails|EnsureAutoresearchWorkspaceWarnsOnPartialScaffoldWithoutOverwriting' -count=1`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If the default `PLANNING.md` scaffold also includes generic `Blocked / Non-Goals` and `Acceptance Criteria` sections, fresh autoresearch workspaces will start with safer planning guardrails instead of leaving those constraints implicit.

## Steps

1. Re-read the current memory and the scaffolded `PLANNING.md` template.
2. Make one bounded change by adding generic blocked/non-goals and acceptance-criteria sections to the default planning scaffold.
3. Run the focused evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- The scaffold should stay generic enough for any autoresearch workspace, even while it becomes more useful for parity-gap planning.
- The parity answer is still expected to be "no"; this cycle only improves default planning guardrails, not launch-time instruction control.
