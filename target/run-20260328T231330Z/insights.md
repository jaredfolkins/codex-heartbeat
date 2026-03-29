# Insights

## What Worked

- Once README exposed the right field, a one-line root-help change was enough to make the new safe task list discoverable from the CLI itself.
- Reusing the focused root-help test kept the cycle bounded while still checking the live help text against the current status surfaces.
- Aligning CLI help after the status and README updates keeps all three operator entry points in sync.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, or benign canary scoring on launch.

## Avoid Next Time

- Do not stop at README updates when built-in CLI help is also part of the normal operator workflow.
- Do not claim Hermes parity from better help text or task-list visibility alone.

## Promising Next Directions

- Consider whether README should show an example `hermes_parity.task_list` payload now that help text and docs both mention it explicitly.
- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
