# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Implement the wrapper-safe phase-1 prompt-profile slice so `codex-heartbeat` can launch Codex with an explicit profile, model, and reasoning effort.
- Primary evaluator: `go test ./cmd/codex-heartbeat -run 'BuildInteractiveArgs|RegisterRunFlags|RunInteractiveCommandPassesLaunchOverrides' -count=1`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: false

## Hypothesis

- If I add wrapper-safe `--profile`, `--model`, and `--model-reasoning-effort` pass-through plus focused tests, the wrapper will move closer to Hermes's launch profile shape without claiming feature parity.

## Steps

1. Read the current memory, local wrapper seam, and upstream Codex CLI syntax.
2. Make one bounded change by adding wrapper-safe launch override flags and observable logging.
3. Add focused tests for the flag registration and child-arg builder.
4. Run the focused evaluator exactly once.
5. Record the result and choose keep, discard, or revert.

## Assumptions

- Upstream Codex CLI already supports `--profile`, `--model`, and config-based `model_reasoning_effort`, so the wrapper can pass them through safely.
- Matching Hermes's full behavior still requires stronger launch-time instruction channels like base/developer instructions, prefill, and canary scoring.
