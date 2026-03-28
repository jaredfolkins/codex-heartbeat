# Program

Objective: Improve the target benchmark without regressing correctness.
Primary evaluator: go test -bench=. ./...
Prompt mode: autoresearch
Council after failures: 2
Checkpoint commits: true

## Constraints

- Keep one benchmark hypothesis per cycle.
- Compare against the same baseline command before and after the change.
- Record the exact benchmark output in `results.md`.

## Notes

- Favor reversible experiments.
- If throughput improves but correctness regresses, revert and record why.
