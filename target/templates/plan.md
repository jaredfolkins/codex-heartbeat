# Plan

- Run: `{{RUN_ID}}`
- Prompt source: `{{PROMPT_SOURCE}}` (`{{PROMPT_SOURCE_PATH}}`)
- Objective: {{OBJECTIVE}}
- Primary evaluator: `{{PRIMARY_EVALUATOR}}`
- Prompt mode: `{{PROMPT_MODE}}`
- Council after failures: {{COUNCIL_AFTER_FAILURES}}
- Checkpoint commits: {{CHECKPOINT_COMMITS}}

## Hypothesis

- {{HYPOTHESIS_PLACEHOLDER}}

## Steps

1. Establish the current baseline.
2. Pick one bounded hypothesis.
3. Make one bounded change.
4. Run the primary evaluator exactly once.
5. Record the result and choose keep, discard, or revert.
