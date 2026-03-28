# Insights

## What Worked

- Recording launch settings in both `latest-context` and the run-start ledger made the current behavior observable without reopening source files.
- Reusing the existing launch summary format kept the artifact change small and consistent with the runtime log text.
- Focused tests made it easy to verify both the saved evidence and the unchanged child launch behavior in one evaluator pass.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, or benign canary scoring on launch.

## Avoid Next Time

- Do not leave program-driven launch settings visible only in code paths when the run artifacts are supposed to hold the session memory.
- Do not claim Hermes parity from better artifact evidence alone.

## Promising Next Directions

- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
- Thread the selected profile/model/effort into any user-facing `status` output so parity reviews do not need to inspect raw artifact files.
