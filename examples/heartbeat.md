Autoresearch override prompt

Use this file only when you want `--prompt` to fully override `program.md`.

- Continue the current objective without restarting from scratch.
- Work in one bounded hypothesis per cycle.
- Reuse one primary evaluator until the current question is answered.
- Update the current run artifacts under `target/run-<timestamp>/`.
- Append a concise result line to `target/results.jsonl`.
- Keep, discard, or revert after each evaluator run.
- Use a 3-agent council only when blocked or clearly stalled.
