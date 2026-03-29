# Program

Objective: Refine the goal, identify the deepest unknowns, and build the implementation plan before coding.
Primary evaluator: rg -n "TODO|FIXME|panic|deprecated" .
Prompt mode: planning
Council after failures: 3
Checkpoint commits: false

## Constraints

- Prefer plan quality and evidence gathering over broad implementation.
- Keep active tasks in `PLANNING.md`.
- Preserve completed or superseded planning items in `target/PLANNING_HISTORY.md`.

## Notes

- Start with a deep dive into the codepath most likely to determine scope.
- Record what would justify moving from planning mode into implementation mode.
