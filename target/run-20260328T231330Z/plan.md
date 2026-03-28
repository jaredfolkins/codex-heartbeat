# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Record the resolved `program.md` launch settings in the run artifacts so the current Hermes-parity answer is backed by saved evidence, not just code inspection.
- Primary evaluator: `go test ./cmd/codex-heartbeat -run 'PromptResolverWritesLaunchSettingsToLatestContext|RecordRunStartWritesEvaluatorToResultsLedger|LoadProgramConfigParsesLaunchOverrides|RegisterRunFlags|RunInteractiveCommandPassesLaunchOverrides' -count=1`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If `latest-context` and the run-start ledger note record the resolved `Profile`, `Model`, and `Model reasoning effort`, the parity answer will be easier to verify from the saved artifacts without changing the actual child Codex launch behavior.

## Steps

1. Re-read the current memory and the current artifact-writing seam around `buildLatestContext()` and `recordRunStart()`.
2. Make one bounded change by threading the resolved launch settings into the saved run artifacts.
3. Run the focused evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- `program.md` remains the authoritative human-edited configuration surface for autoresearch runs.
- The parity answer is still expected to be "no"; this cycle is about making that answer observable in artifacts rather than expanding the actual feature surface.
