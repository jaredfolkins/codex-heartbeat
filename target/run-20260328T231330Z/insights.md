# Insights

## What Worked

- `PLANNING.md` is still the best place to express the requested `[ ]` implementation backlog once the status/help/docs surfaces already agree.
- A simple grep evaluator was enough to keep the planning cycle bounded while still checking that the `Phase 1 Recommendation` section carried the new traceability task.
- Source-grounding details such as `review_basis` belong in the near-term plan too, not only in broad backlog or status/docs surfaces.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, benign canary scoring, or Hermes-style delegated cross-review on launch.

## Avoid Next Time

- Do not let the planning backlog drift behind the already-documented parity and traceability surfaces.
- Do not claim Hermes parity from a more complete task list alone.

## Promising Next Directions

- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Decide whether the existing fallback council needs a more Hermes-like first-class delegated review surface for benign evaluator comparisons.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
