# Program

Objective: Reproduce and fix the current failing behavior with the smallest safe change.
Primary evaluator: go test ./...
Prompt mode: autoresearch
Council after failures: 3
Checkpoint commits: true

## Constraints

- Keep the hypothesis narrow and falsifiable.
- Reuse the same evaluator until it stops answering the current question.
- Do not expand scope just because a test passes.

## Notes

- Prefer baseline-first debugging.
- Update `target/run-<timestamp>/` artifacts every cycle.
