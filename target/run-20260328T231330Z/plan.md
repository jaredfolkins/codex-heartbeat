# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Tighten the source-grounded `[ ]` backlog in `PLANNING.md` so it explicitly covers how parity claims stay auditable against the reviewed Hermes materials.
- Primary evaluator: `rg -n "review_basis|source-traceability|traceable to reviewed source material|source-grounded" PLANNING.md`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If `PLANNING.md` explicitly includes source-traceability items around `review_basis`, the source-grounded implementation backlog will better cover how parity claims stay auditable against the reviewed Hermes materials.

## Steps

1. Re-read the current memory and the existing planning backlog.
2. Make one bounded change by adding source-traceability items to `PLANNING.md`.
3. Run the focused planning evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- The parity answer is still expected to be "no"; this cycle only improves the implementation backlog.
- The planning change must stay in the safe research-workflow / observability lane and must not suggest bypass or jailbreak behavior.
