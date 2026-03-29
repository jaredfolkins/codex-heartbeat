# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Ground the `status.hermes_parity` answer more explicitly in the reviewed sources by exposing a small `review_basis` field.
- Primary evaluator: `go test ./cmd/codex-heartbeat -run 'StatusCommandIncludesHermesParityGap|StatusCommandIncludesProgramLaunchSettings' -count=1`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If `status.hermes_parity` includes a small `review_basis` field naming the reviewed Hermes repo and X post, operators will be able to see that the current non-parity answer is grounded in those sources instead of only in local wording.

## Steps

1. Re-read the current memory, the current parity surface, and the reviewed-source notes.
2. Make one bounded change by adding a `review_basis` field to `hermes_parity`.
3. Run the focused status evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- The parity answer is still expected to be "no"; this cycle only makes the status answer more traceable to the reviewed sources.
- The new field must stay in the safe research-workflow / observability lane and must not suggest bypass or jailbreak behavior.
