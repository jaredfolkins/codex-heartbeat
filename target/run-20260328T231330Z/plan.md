# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Make the source-grounded `[ ]` backlog in `PLANNING.md` explicitly cite the two reviewed links, not just summarize them.
- Primary evaluator: `rg -n "^### Review Basis|2037294903814738261|github.com/nousresearch/hermes-agent|cross-review|launch-time instruction control" PLANNING.md`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If `PLANNING.md` explicitly includes a `Review Basis` section with the two reviewed links, the source-grounded task list will be visibly anchored to the exact materials the user asked to review.

## Steps

1. Re-read the current memory and the existing planning backlog.
2. Make one bounded change by adding a `Review Basis` section to `PLANNING.md`.
3. Run the focused planning evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- The parity answer is still expected to be "no"; this cycle only improves the implementation backlog.
- The planning change must stay in the safe research-workflow / observability lane and must not suggest bypass or jailbreak behavior.
