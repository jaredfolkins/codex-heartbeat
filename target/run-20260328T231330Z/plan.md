# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Document the new `status` surfaces so operators can discover `launch_settings` and `hermes_parity` without inspecting source or raw JSON by guesswork.
- Primary evaluator: `rg -n "^Inspect the stored session:|status --workdir|launch_settings|hermes_parity|not equivalent to Hermes Agent's" README.md`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If the README explains that `codex-heartbeat status` exposes `launch_settings` and `hermes_parity`, the current “not the same as Hermes” answer will be easier for operators to verify from the documented workflow.

## Steps

1. Re-read the current memory and the README sections around `status` and launch profiles.
2. Make one bounded change by documenting the new `status` fields.
3. Run the focused README evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- `program.md` remains the authoritative human-edited configuration surface for autoresearch runs.
- The parity answer is still expected to be "no"; this cycle is about documentation, not changing the feature surface.
