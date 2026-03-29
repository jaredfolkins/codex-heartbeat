# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Mention the new `hermes_parity.task_list` surface in built-in CLI help so operators can discover the safe next-step checklist even before reading the README.
- Primary evaluator: `go test ./cmd/codex-heartbeat -run 'RootUsageMentionsStatusSurfaces|StatusCommandIncludesHermesParityGap|StatusCommandIncludesProgramLaunchSettings' -count=1`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If `printRootUsage()` mentions that `hermes_parity` includes the safe `task_list`, operators will be able to discover the parity next-step surface directly from CLI help instead of only from README text or raw JSON output.

## Steps

1. Re-read the current memory and the built-in root usage text.
2. Make one bounded change by mentioning the `task_list` field inside the `status` help line.
3. Run the focused evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- The parity answer is still expected to be "no"; this cycle only improves CLI-help discoverability.
- The help text must keep the new task list in the safe prompt-profile / observability lane and must not suggest bypass or jailbreak behavior.
