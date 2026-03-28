# Plan

- Hypothesis: A short artifact-first README section can improve heartbeat triage and contributor verification without freezing fragile detector internals into the docs.

## Steps

1. Re-read `README.md`, `AGENTS.md`, and prior `target/*/insights.md` artifacts.
2. Run the required 3-agent council with the requested `gpt-5.3-codex-spark` model and choose a docs-only plan.
3. Add one concise README section with an operator quick-triage path and a contributor verification path.
4. Validate the exact test commands referenced by the docs and run `go test ./...`.
5. Update run memory, re-review `AGENTS.md`, and create a save-point commit.

## Assumptions

- The highest-value docs gap is not setup or prompt precedence, but how to triage heartbeat behavior and verify it without guessing from internals.
- The docs should stay artifact-first and avoid turning current detector heuristics into a permanent public contract.
- Interrupted placeholder run artifacts from this session should be cleaned up rather than preserved as noise.
