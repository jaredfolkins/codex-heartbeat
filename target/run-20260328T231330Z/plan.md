# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Mention the new `status` surfaces in built-in CLI help so operators can discover them even before reading the README.
- Primary evaluator: `go test ./cmd/codex-heartbeat -run 'RootUsageMentionsStatusSurfaces|StatusCommandIncludesHermesParityGap|StatusCommandIncludesProgramLaunchSettings|PromptResolverWritesLaunchSettingsToLatestContext|RecordRunStartWritesEvaluatorToResultsLedger|LoadProgramConfigParsesLaunchOverrides|RegisterRunFlags|RunInteractiveCommandPassesLaunchOverrides' -count=1`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If `printRootUsage()` mentions that `status` exposes `launch_settings` and `hermes_parity`, operators will be able to discover the current parity-explanation surface directly from CLI help without affecting runtime behavior.

## Steps

1. Re-read the current memory and the built-in root usage text.
2. Make one bounded change by adding a short help line about the `status` surfaces.
3. Run the focused evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- `program.md` remains the authoritative human-edited configuration surface for autoresearch runs.
- The parity answer is still expected to be "no"; this cycle is about CLI help discoverability, not changing the feature surface.
