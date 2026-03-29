# Insights

## What Worked

- A small scaffold change was enough to add safer default planning guardrails without changing runtime behavior.
- Focused scaffold tests are a good place to lock in planning ergonomics because they avoid coupling these changes to live session behavior.
- Mirroring the live workspace's blocked/non-goals and acceptance-criteria shape in the default scaffold makes future autoresearch workspaces safer to steer from the first run.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, or benign canary scoring on launch.

## Avoid Next Time

- Do not assume the default scaffold is “good enough” just because the current repo has a richer hand-edited planning file.
- Do not claim Hermes parity from better planning guardrails or checklist scaffolding alone.

## Promising Next Directions

- Consider whether the default scaffold should also seed a short parity-gap or transport-boundary note without becoming too domain-specific.
- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
