# Insights

## What Worked

- Putting the Hermes parity gap into `status` made the current “no” answer explicit from the main operator command instead of forcing a doc/artifact cross-reference.
- Reusing a small structured `hermes_parity` object kept the change narrow and easy to test.
- Focused tests made it easy to verify the new parity surface alongside the existing launch-setting and artifact evidence in one evaluator pass.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, or benign canary scoring on launch.

## Avoid Next Time

- Do not rely on prose alone for parity answers when the operator-facing command can return a structured status block.
- Do not claim Hermes parity from better status visibility alone.

## Promising Next Directions

- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
- If the status surface grows further, consider grouping prompt-profile and parity fields under a single structured feature-status section.
