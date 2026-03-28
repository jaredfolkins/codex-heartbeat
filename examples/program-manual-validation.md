# Program

Objective: Prepare the next candidate fix and the exact manual validation steps.
Primary evaluator: make manual-lab-up
Prompt mode: manual-test-first
Council after failures: 2
Checkpoint commits: false

## Constraints

- Stop before the final human gate.
- Produce the next patch and the exact manual validation checklist.
- Record what the human should verify next.

## Notes

- This mode is guidance, not a hard safety boundary.
- Keep the validation instructions concrete and reusable.
