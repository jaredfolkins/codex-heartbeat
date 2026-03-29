# Results

- Status: success
- Council triggered at start: false
- Primary evaluator: `rg -n "^Inspect the stored session:|status --workdir|launch_settings|hermes_parity|task_list|claim_rule|review_basis|delegated cross-review|not equivalent to Hermes Agent's" README.md`

## Observable Signals

- The README parity section now says that `hermes_parity` includes a `review_basis` list pointing at the reviewed Hermes sources.
- The focused evaluator passed for the updated README status and parity wording.
- Operators can now discover the source-grounded parity explanation from README as well as raw `status` JSON.
- The function still does not appear to be the same as Hermes Agent because the wrapper still lacks stronger launch-time instruction control, ephemeral prefill, benign canary scoring, and a Hermes-style delegated cross-review workflow.

## Disposition

- keep
