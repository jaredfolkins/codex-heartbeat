# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Expose a concrete safe Hermes parity task list in `codex-heartbeat status` so the current “not equivalent” answer includes operator-visible next steps instead of only missing capabilities.
- Primary evaluator: `go test ./cmd/codex-heartbeat -run 'StatusCommandIncludesHermesParityGap|StatusCommandIncludesProgramLaunchSettings' -count=1`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If `status.hermes_parity` includes a safe `task_list` alongside `missing`, the repo will answer the current parity question with concrete next steps instead of only a negative capability list.

## Steps

1. Re-read the current memory and the current `hermes_parity` status surface.
2. Make one bounded change by adding a safe `task_list` field to `hermes_parity`.
3. Run the focused evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- The parity answer is still expected to be "no"; this cycle only improves the operator-facing explanation of the gap.
- The task list must stay in the safe prompt-profile / observability lane and must not suggest bypass or jailbreak behavior.
