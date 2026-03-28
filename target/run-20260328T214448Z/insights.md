# Insights

## What Worked

- Reviewing the upstream TUI snapshots again kept the change narrow and tied to real Codex screen states instead of guessed phrases.
- Explicit post-turn markers are a safe place to reduce ambiguity without relaxing live working detection.
- A rollout tiebreaker that only runs after `screenStateAmbiguous` keeps the existing scheduler semantics intact while still using stronger persisted signals.
- `task_complete`, `context_compacted`, `task_started`, and unmatched `function_call` entries are high-signal rollout markers for post-turn versus in-flight states.

## What Failed

- Persisted `.codex` state does not give a simple on-disk live thread-status flag; the clean status model lives in app-server notifications instead.
- Go test cases that call `t.Setenv(...)` cannot also be marked `t.Parallel()`.

## Avoid Next Time

- Do not assume `.codex/state_*.sqlite` has a direct `status` column for heartbeat decisions.
- Do not combine environment mutation helpers like `t.Setenv(...)` with `t.Parallel()` in new tests.

## Promising Next Directions

- If ambiguity remains a problem, extend the rollout tiebreaker with additional high-signal persisted events while keeping it strictly secondary to the live screen classifier.
