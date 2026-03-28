# Plan

- Hypothesis: A prompt resolver plus bounded autoresearch artifacts can satisfy the AGENTS.md checklist without regressing existing CLI behavior.

## Steps

1. Read the current CLI, prompt flow, and tests.
2. Run a 3-agent council and pick the winning plan.
3. Add prompt precedence across `--prompt`, `program.md`, and the embedded fallback template.
4. Add bounded `target/` artifacts, latest-context generation, prior-insight ingestion, and ledger helpers.
5. Add `init` scaffolding, example programs, README docs, and tests.
6. Validate with `go test ./...` and `go test -race ./...`.

## Assumptions

- The wrapper should stay a prompt-injection tool, not a full autonomous executor.
- `program.md` should be human-authored and re-read on each prompt emission.
- The council fallback can be driven by recent `target/results.jsonl` failure streaks.
