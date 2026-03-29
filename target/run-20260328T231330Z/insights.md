# Insights

## What Worked

- After adding `launch_settings` and `hermes_parity` to `status`, a short README update was enough to make the new operator path discoverable.
- A text-structure evaluator was sufficient for this cycle because the runtime behavior had already been covered by the earlier focused tests.
- Keeping the docs change isolated avoided reopening the implementation just to align the user-facing workflow.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, or benign canary scoring on launch.

## Avoid Next Time

- Do not add operator-facing surfaces without documenting where to find them in the normal workflow.
- Do not treat documentation alignment as optional once the status surface becomes the primary explanation path.

## Promising Next Directions

- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
- If the status surface keeps growing, consider a short dedicated README subsection that shows an example JSON payload and how to interpret it.
