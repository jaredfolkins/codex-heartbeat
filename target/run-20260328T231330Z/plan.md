# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Tighten the source-grounded `[ ]` backlog in `PLANNING.md` so it explicitly covers the reviewed Hermes-style delegated cross-review workflow.
- Primary evaluator: `rg -n "delegated cross-review|Hermes Parity Gap|Phase 1 Recommendation|Task List" PLANNING.md`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If `PLANNING.md` explicitly includes Hermes-style delegated cross-review items in the task list, phase-1 recommendation, and parity-gap sections, the source-grounded implementation backlog will better match the reviewed multi-LLM research workflow.

## Steps

1. Re-read the current memory and the existing planning backlog.
2. Make one bounded change by adding delegated cross-review items to `PLANNING.md`.
3. Run the focused planning evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- The parity answer is still expected to be "no"; this cycle only improves the implementation backlog.
- The planning change must stay in the safe research-workflow / observability lane and must not suggest bypass or jailbreak behavior.
