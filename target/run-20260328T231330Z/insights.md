# Insights

## What Worked

- README is a good home for parity-surface clarifications when the runtime behavior is already correct and the next problem is operator discoverability.
- A single grep-based evaluator was enough to keep the docs cycle bounded while still checking the key parity terms together.
- Once `claim_rule` and delegated cross-review existed in `status`, mirroring them in the README made the Hermes comparison easier to scan.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, benign canary scoring, or Hermes-style delegated cross-review on launch.

## Avoid Next Time

- Do not update docs in a way that drifts from the actual `status` JSON surface.
- Do not claim Hermes parity from clearer documentation alone.

## Promising Next Directions

- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Decide whether the existing fallback council needs a more Hermes-like first-class delegated review surface for benign evaluator comparisons.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
