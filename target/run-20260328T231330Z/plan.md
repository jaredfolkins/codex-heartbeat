# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Make the Hermes parity gap explicit in the task list so the repo records exactly why the current wrapper is still not the same feature.
- Primary evaluator: `rg -n "^### Hermes Parity Gap|stronger launch-time instruction channel|ephemeral prefill|harmless canary-scoring harness|parity claim rule" PLANNING.md`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If I add an explicit Hermes parity-gap checklist to `PLANNING.md`, the repo can answer the current evaluator with a concrete delta instead of only a freeform explanation.

## Steps

1. Read the current memory and existing planning backlog.
2. Make one bounded change by adding an explicit Hermes parity-gap checklist.
3. Run the planning evaluator exactly once.
5. Record the result and choose keep, discard, or revert.

## Assumptions

- The current backlog already contains most of the raw ingredients; the missing piece is an explicit parity checklist.
- A clearer gap list is useful enough to justify a small follow-up save point.
