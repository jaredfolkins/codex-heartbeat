# Insights

## What Worked

- Once `status` exposed the right fields, a one-line root-help update was enough to make the feature discoverable from the CLI itself.
- Reusing the existing focused evaluator bundle kept the cycle bounded while still checking the new help text against the live status surfaces.
- Aligning CLI help after README kept the user-facing explanation path consistent across both entry points.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, or benign canary scoring on launch.

## Avoid Next Time

- Do not stop at README updates when built-in CLI help is also part of the normal operator workflow.
- Do not claim Hermes parity from better help text or discoverability alone.

## Promising Next Directions

- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
- If the status surface keeps growing, consider a short dedicated README subsection that shows an example JSON payload and how to interpret it.
