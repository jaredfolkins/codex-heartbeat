# Insights

## What Worked

- Root CLI help is a good final parity-discoverability surface once README and raw `status` JSON already agree.
- Reusing the focused root-help/status test kept the cycle bounded while still rechecking the parity and launch-setting surfaces together.
- Once `review_basis` existed in `status` and README, mirroring it in root help completed the main operator-visible parity explanation path.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, benign canary scoring, or Hermes-style delegated cross-review on launch.

## Avoid Next Time

- Do not update help text in a way that drifts from the actual `status` JSON surface.
- Do not claim Hermes parity from more complete help text alone.

## Promising Next Directions

- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Decide whether the existing fallback council needs a more Hermes-like first-class delegated review surface for benign evaluator comparisons.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
