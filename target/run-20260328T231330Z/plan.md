# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Expose the current Hermes parity gap directly in `codex-heartbeat status` so the “same as Hermes or not?” answer is explicit from one command.
- Primary evaluator: `go test ./cmd/codex-heartbeat -run 'StatusCommandIncludesHermesParityGap|StatusCommandIncludesProgramLaunchSettings|PromptResolverWritesLaunchSettingsToLatestContext|RecordRunStartWritesEvaluatorToResultsLedger|LoadProgramConfigParsesLaunchOverrides|RegisterRunFlags|RunInteractiveCommandPassesLaunchOverrides' -count=1`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If `codex-heartbeat status` includes a `hermes_parity` block with `equivalent=false` and the concrete missing capabilities, the current parity answer will be obvious from the operator-facing surface without changing any launch behavior.

## Steps

1. Re-read the current memory and the current `status` JSON surface.
2. Make one bounded change by threading the Hermes parity gap into `status` output.
3. Run the focused evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- `program.md` remains the authoritative human-edited configuration surface for autoresearch runs.
- The parity answer is still expected to be "no"; this cycle is about making that answer explicit in `status`, not expanding the actual feature surface.
