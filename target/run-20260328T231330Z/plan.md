# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Document the phase-1 launch-profile support and the remaining Hermes gap so the README matches the code and the planning backlog.
- Primary evaluator: `rg -n "^## Launch Profiles|--profile NAME|--model-reasoning-effort LEVEL|not equivalent to Hermes Agent's|phase-1 prompt-profile feature" README.md`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If I document the new launch flags and the remaining non-parity in `README.md`, the repo's operator guidance will match the implemented phase-1 feature and the current "not the same yet" answer.

## Steps

1. Read the current memory, README, and the current launch-profile seam.
2. Make one bounded change by documenting the wrapper-safe launch-profile support and non-parity in `README.md`.
3. Run the README evaluator exactly once.
5. Record the result and choose keep, discard, or revert.

## Assumptions

- The code already supports the wrapper-safe launch flags, but the README does not describe them yet.
- A short doc update is enough to make the current phase-1 state easier to understand without expanding scope.
