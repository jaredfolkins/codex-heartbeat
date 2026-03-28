# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Expose the resolved `program.md` launch settings in `codex-heartbeat status` so the current Hermes-parity answer is visible from the main operator command.
- Primary evaluator: `go test ./cmd/codex-heartbeat -run 'StatusCommandIncludesProgramLaunchSettings|PromptResolverWritesLaunchSettingsToLatestContext|RecordRunStartWritesEvaluatorToResultsLedger|LoadProgramConfigParsesLaunchOverrides|RegisterRunFlags|RunInteractiveCommandPassesLaunchOverrides' -count=1`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If `codex-heartbeat status` includes the resolved `Profile`, `Model`, and `Model reasoning effort`, operators will be able to verify the program-driven launch configuration without digging through artifact files, while the actual child launch behavior stays the same.

## Steps

1. Re-read the current memory and the `status` command surface.
2. Make one bounded change by threading the resolved launch settings into `status` JSON output.
3. Run the focused evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- `program.md` remains the authoritative human-edited configuration surface for autoresearch runs.
- The parity answer is still expected to be "no"; this cycle is about making that answer observable from `status` rather than expanding the actual feature surface.
