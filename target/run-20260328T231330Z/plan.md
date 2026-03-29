# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Ground `status.hermes_parity` in the reviewed Hermes references by adding the remaining delegated cross-review workflow gap to the current safe parity surface.
- Primary evaluator: `go test ./cmd/codex-heartbeat -run 'StatusCommandIncludesHermesParityGap|StatusCommandIncludesProgramLaunchSettings' -count=1`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If `status.hermes_parity` explicitly includes the Hermes-style delegated cross-review workflow gap surfaced by the reviewed Hermes sources, the current parity answer will better match what Hermes Agent actually offers for research workflows.

## Steps

1. Re-read the current memory plus the referenced Hermes materials.
2. Make one bounded change by adding the reviewed delegated cross-review gap to `hermes_parity`.
3. Run the focused evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- The parity answer is still expected to be "no"; this cycle only tightens the operator-facing explanation of the gap.
- The new gap item must stay in the safe research-workflow / observability lane and must not suggest bypass or jailbreak behavior.
