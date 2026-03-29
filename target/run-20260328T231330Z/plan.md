# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Document the new `hermes_parity.task_list` status surface so operators can discover the safe next-step checklist from the normal README workflow.
- Primary evaluator: `rg -n "^Inspect the stored session:|status --workdir|launch_settings|hermes_parity|task_list|not equivalent to Hermes Agent's" README.md`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If the README explains that `status.hermes_parity` includes a safe `task_list`, operators will be able to discover the new parity next-step surface from the documented workflow instead of only from raw JSON output.

## Steps

1. Re-read the current memory and the documented `status` workflow in the README.
2. Make one bounded change by documenting the `task_list` field inside `hermes_parity`.
3. Run the focused evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- The parity answer is still expected to be "no"; this cycle only improves operator-facing documentation.
- The docs must keep the new task list in the safe prompt-profile / observability lane and must not suggest bypass or jailbreak behavior.
