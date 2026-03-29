# Insights

## What Worked

- README is a good follow-up surface once the status payload already carries the right source-grounding data.
- A single grep-based evaluator was enough to keep the docs cycle bounded while still checking the key parity terms together.
- Once `review_basis` existed in `status`, mirroring it in README completed the main operator-visible parity explanation path.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, benign canary scoring, or Hermes-style delegated cross-review on launch.

## Avoid Next Time

- Do not update docs in a way that drifts from the actual `status` JSON surface.
- Do not claim Hermes parity from more complete parity docs alone.

## Promising Next Directions

- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Decide whether the existing fallback council needs a more Hermes-like first-class delegated review surface for benign evaluator comparisons.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
