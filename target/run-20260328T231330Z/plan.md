# Plan

- Run: `run-20260328T231330Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Move launch profile selection into `program.md` metadata so the autoresearch program owns model/profile/effort instead of top-level wrapper flags.
- Primary evaluator: `go test ./cmd/codex-heartbeat -run 'LoadProgramConfigParsesLaunchOverrides|RegisterRunFlags|RunInteractiveCommandPassesLaunchOverrides' -count=1`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: true

## Hypothesis

- If `Profile`, `Model`, and `Model reasoning effort` move into `program.md` metadata and the wrapper stops exposing them as top-level flags, the launch surface will better match the autoresearch contract without changing the downstream child Codex CLI args.

## Steps

1. Re-read the current memory, `program.md`, and the current launch-profile seam.
2. Make one bounded change by moving launch selection into `program.md` parsing and removing the redundant wrapper flags.
3. Run the focused evaluator exactly once.
4. Record the result and choose keep, discard, or revert.

## Assumptions

- `program.md` is the authoritative human-edited configuration surface for autoresearch runs.
- The child Codex CLI still needs the same `--profile`, `--model`, and `--config model_reasoning_effort=...` args once the wrapper resolves them from `program.md`.
