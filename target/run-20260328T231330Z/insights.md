# Insights

## What Worked

- A small `status` payload change was enough to make the parity answer more auditable without changing runtime behavior.
- Reusing the focused status test kept the cycle bounded while still rechecking both the parity surface and the launch-settings surface together.
- A `review_basis` field is a clean way to connect local parity wording back to the actual reviewed sources.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, benign canary scoring, or Hermes-style delegated cross-review on launch.

## Avoid Next Time

- Do not add source-grounding fields that are not actually reflected in the reviewed materials.
- Do not claim Hermes parity from more traceable status output alone.

## Promising Next Directions

- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Decide whether the existing fallback council needs a more Hermes-like first-class delegated review surface for benign evaluator comparisons.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
