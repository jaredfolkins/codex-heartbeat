# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Align root CLI help with the current safe `status.hermes_parity` surface by exposing `claim_rule` and the Hermes-style review gap in the built-in help text.
- Primary evaluator: `go test ./cmd/codex-heartbeat -run 'RootUsageMentionsStatusSurfaces|StatusCommandIncludesHermesParityGap|StatusCommandIncludesProgramLaunchSettings' -count=1`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If root CLI help explicitly mentions the safe parity `claim_rule` and the Hermes-style review gap, operators will be able to discover the fuller non-parity explanation from help text without opening README or raw `status` JSON.

## Steps

1. Re-read the current memory, the current parity surface, and the root-help text.
2. Make one bounded change by documenting `claim_rule` and the Hermes-style review gap in root help.
3. Run the focused root-help evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- The parity answer is still expected to be "no"; this cycle only aligns help text with the existing safe parity surface.
- The help-text change must stay in the safe research-workflow / observability lane and must not suggest bypass or jailbreak behavior.
