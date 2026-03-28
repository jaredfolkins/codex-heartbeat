# Plan

- Hypothesis: A layered replay test suite can verify the completed AGENTS.md implementation and harden heartbeat/council behavior without further runtime changes.

## Steps

1. Re-read `AGENTS.md` and prior `target/*/insights.md` artifacts.
2. Run the required 3-agent council and challenge the assumption that the checklist is truly complete.
3. Select the winning plan using feasibility, testability, real-world success, and execution clarity.
4. Add layered tests for screen classification, scheduler timelines, thin end-to-end replay, and council-threshold robustness.
5. Run focused tests, then `go test ./...` and `go test -race ./...`.
6. Update run memory, re-review `AGENTS.md`, and save a commit.

## Assumptions

- The AGENTS feature checklist is already implemented, so the remaining justified work is verification depth rather than new runtime behavior.
- Layered tests are better than a monolithic replay harness because they localize failures quickly.
- Runtime behavior should change only if the new coverage exposes a real bug.
