# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Make the source-grounded `[ ]` backlog in `PLANNING.md` explicitly cover Hermes-style global identity-file semantics.
- Primary evaluator: `rg -n "global-identity file semantics|SOUL.md lives only in HERMES_HOME|occupies identity slot `#1`|per-user identity file exists outside repo context|fallback identity applies when that file is absent or empty" PLANNING.md`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If `PLANNING.md` explicitly includes Hermes-style global identity-file items, the implementation backlog will better match the reviewed Hermes operator workflow instead of leaving the per-user identity layer implicit.

## Steps

1. Re-read the current memory and the existing planning backlog.
2. Make one bounded change by adding Hermes-style global identity-file items to `PLANNING.md`.
3. Run the focused planning evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- The parity answer is still expected to be "no"; this cycle only improves the implementation backlog.
- The planning change must stay in the safe research-workflow / observability lane and must not suggest bypass or jailbreak behavior.
