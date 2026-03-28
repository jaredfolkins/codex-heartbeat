# Insights

## What Worked

- Splitting the work into prompt-resolution first and artifact/memory wiring second kept the refactor manageable.
- A dedicated `autoresearch.go` file kept the new behavior out of the hottest CLI paths.
- The results ledger is a practical place to drive council fallback thresholds without changing scheduler semantics.

## What Failed

- The first integration pass left one deterministic screen-log filename bug, which the test suite caught immediately.

## Avoid Next Time

- Do not write run artifacts outside the workspace lock.
- Do not let cached explicit prompts leak into `program.md` mode when `--prompt` is not set.

## Promising Next Directions

- Run `codex-heartbeat init --workdir <repo>` against a real target workspace and exercise the new prompt flow end to end.
- Consider adding a small integration test around a fake `program.md` plus `results.jsonl` streak to assert the assembled council instruction text.
