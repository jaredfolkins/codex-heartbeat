# Plan

- Run: `run-20260328T214448Z`
- Prompt source: `program_md` (`/Users/jf/src/jf/codex-heartbeat/program.md`)
- Objective: Reduce ambiguous heartbeat classifications by using `.codex` rollout events as a tiebreaker only when the screen classifier cannot decide.
- Primary evaluator: `go test ./cmd/codex-heartbeat -run 'Screen|Replay' -count=1`
- Prompt mode: `autoresearch`
- Council after failures: 3
- Checkpoint commits: false

## Hypothesis

- Consulting the current session rollout only for `ambiguous` screens will reduce false heartbeat stalls without relaxing the existing idle/quiet scheduler thresholds.

## Steps

1. Review the upstream Codex CLI rollout format and local `.codex/sessions/.../rollout-*.jsonl` records to identify persisted turn-complete and in-flight signals.
2. Add one bounded ambiguous-only tiebreaker that resolves the active rollout path through `CODEX_HOME`, inspects the recent rollout tail, and promotes `task_complete` / `context_compacted` to idle plus `task_started` / pending `function_call` to working.
3. Validate with `go test ./cmd/codex-heartbeat -run 'Screen|Replay' -count=1`.
4. Record the result, including any test-harness deviations discovered during validation.

## Assumptions

- Codex rollout filenames continue to include the full session ID, with `session_meta` available as a fallback lookup path.
- Persisted rollout events are trustworthy enough to break ties only after the screen classifier has already returned `ambiguous`.
