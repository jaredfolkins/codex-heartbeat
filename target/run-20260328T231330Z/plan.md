# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Make the source-grounded `[ ]` backlog in `PLANNING.md` explicitly cover Hermes-style profile-persistence scope for named bundles.
- Primary evaluator: `rg -n "persistence rule|profile persistence scope|one-session override|persisted default for future sessions|session-local, repo-default" PLANNING.md`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If `PLANNING.md` explicitly includes Hermes-style profile-persistence-scope items, the implementation backlog will better match the reviewed Hermes operator workflow instead of leaving one-session overrides versus persisted defaults ambiguous.

## Steps

1. Re-read the current memory and the existing planning backlog.
2. Make one bounded change by adding Hermes-style profile-persistence-scope items to `PLANNING.md`.
3. Run the focused planning evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- The parity answer is still expected to be "no"; this cycle only improves the implementation backlog.
- The planning change must stay in the safe research-workflow / observability lane and must not suggest bypass or jailbreak behavior.
