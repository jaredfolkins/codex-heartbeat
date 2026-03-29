# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Make the source-grounded `[ ]` backlog in `PLANNING.md` explicitly cover Hermes-style launch-time skill/toolset loading within named bundles.
- Primary evaluator: `rg -n "launch-time skill and toolset loading|bundle-to-tooling semantics|preloads toolsets or skills|selected named bundle also loads extra toolsets or skills|named bundles may preload toolsets or skills" PLANNING.md`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If `PLANNING.md` explicitly includes Hermes-style bundle-to-tooling items, the implementation backlog will better match the reviewed Hermes operator workflow instead of treating named bundles as prompt-only state.

## Steps

1. Re-read the current memory and the existing planning backlog.
2. Make one bounded change by adding Hermes-style bundle-to-tooling items to `PLANNING.md`.
3. Run the focused planning evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- The parity answer is still expected to be "no"; this cycle only improves the implementation backlog.
- The planning change must stay in the safe research-workflow / observability lane and must not suggest bypass or jailbreak behavior.
