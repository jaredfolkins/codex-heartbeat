# Insights

## What Worked

- Once `status` exposed the right field, a one-line README change was enough to make the new safe task list discoverable from the documented operator workflow.
- A focused `rg` evaluator is a good fit for docs-alignment cycles because it verifies the exact user-facing strings without widening into runtime checks.
- Keeping README in sync right after the `status` surface change prevents the docs from lagging behind the operator-visible JSON.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, or benign canary scoring on launch.

## Avoid Next Time

- Do not stop at status JSON changes when the README still describes the older surface.
- Do not claim Hermes parity from better docs or task-list visibility alone.

## Promising Next Directions

- Consider whether README should show an example `hermes_parity.task_list` payload now that the docs mention it explicitly.
- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
