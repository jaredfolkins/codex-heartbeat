# Insights

## What Worked

- Moving launch selection into `program.md` matched the autoresearch contract better than keeping those settings on the wrapper CLI.
- Extending `programConfig` let the source of truth move without changing the downstream child-arg builder.
- Focused tests made it easy to verify the intended surface change: metadata present, wrapper flags absent, child args unchanged.
- Updating the README in the same cycle kept operator guidance aligned with the new configuration model.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, or benign canary scoring on launch.

## Avoid Next Time

- Do not add top-level wrapper flags for settings that belong in the human-edited autoresearch program.
- Do not confuse child Codex launch args with the wrapper's own public configuration surface.
- Do not claim Hermes parity from profile/model/effort placement alone.

## Promising Next Directions

- Thread the selected profile/model/effort into execution artifacts so each run records exactly which `program.md` launch metadata was used.
- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
