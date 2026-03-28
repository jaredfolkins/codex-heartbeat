# Insights

## What Worked

- Surfacing launch settings through `status` made the current configuration easier to inspect from the normal operator path, not just from memory artifacts.
- Reusing the existing `launchOverrides` shape kept the change small and consistent across status, artifacts, and runtime logs.
- Focused tests made it easy to verify the new `status` surface and the unchanged child launch behavior in one evaluator pass.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, or benign canary scoring on launch.

## Avoid Next Time

- Do not leave important program-driven configuration visible only in internal artifacts when the operator-facing `status` command can expose it directly.
- Do not claim Hermes parity from better observability alone.

## Promising Next Directions

- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
- Expose the remaining Hermes parity gap itself through a user-facing surface so the current "no" answer does not require correlating README, PLANNING, and artifacts by hand.
