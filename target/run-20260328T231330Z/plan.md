# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Seed fresh autoresearch workspaces with an explicit checkbox task list in scaffolded `PLANNING.md` so the safe Hermes-inspired planning shape appears by default.
- Primary evaluator: `go test ./cmd/codex-heartbeat -run 'EnsureAutoresearchWorkspaceScaffoldsWorkspace|EnsureAutoresearchWorkspaceSeedsPlanningTaskList|EnsureAutoresearchWorkspaceWarnsOnPartialScaffoldWithoutOverwriting' -count=1`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If the default `PLANNING.md` scaffold includes a checkbox task-list section, fresh autoresearch workspaces will start with actionable `[ ]` planning items instead of only prose headings, making the current safe Hermes-inspired implementation backlog easier to instantiate.

## Steps

1. Re-read the current memory and the scaffolded `PLANNING.md` template.
2. Make one bounded change by adding a checkbox task-list section to the default planning scaffold.
3. Run the focused evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- The scaffold should stay generic enough for any autoresearch workspace, even while it becomes more useful for parity-gap planning.
- The parity answer is still expected to be "no"; this cycle only improves default task-list scaffolding, not launch-time instruction control.
