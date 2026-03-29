# Insights

## What Worked

- A small `status` payload change was enough to turn the parity answer into a concrete safe task list without changing runtime behavior.
- Reusing the existing focused status tests kept the cycle bounded while still validating both the parity surface and the launch-settings surface together.
- Exposing next steps directly in `status` is a better operator fit than forcing users to infer them from a raw `missing` array alone.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, or benign canary scoring on launch.

## Avoid Next Time

- Do not assume a raw capability gap list is enough when the objective is “what should we do next?” rather than only “what is missing?”
- Do not claim Hermes parity from better task-list visibility or operator guidance alone.

## Promising Next Directions

- Consider whether README should show an example `hermes_parity.task_list` payload now that `status` exposes it directly.
- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
