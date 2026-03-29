# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Align the README parity explanation with the current safe `status.hermes_parity` surface by documenting `review_basis`.
- Primary evaluator: `rg -n "^Inspect the stored session:|status --workdir|launch_settings|hermes_parity|task_list|claim_rule|review_basis|delegated cross-review|not equivalent to Hermes Agent's" README.md`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If the README parity section explicitly includes `review_basis`, operators will be able to see the current safe parity answer is source-grounded without inspecting the raw `status` JSON.

## Steps

1. Re-read the current memory, the current parity surface, and the README parity section.
2. Make one bounded change by documenting `review_basis` in README.
3. Run the focused README evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- The parity answer is still expected to be "no"; this cycle only aligns docs with the existing safe parity surface.
- The docs change must stay in the safe research-workflow / observability lane and must not suggest bypass or jailbreak behavior.
