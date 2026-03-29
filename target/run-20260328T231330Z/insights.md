# Insights

## What Worked

- A small scaffold change was enough to turn the generic planning template into a real checkbox backlog without changing runtime behavior.
- A focused scaffold test is a good place to lock in planning ergonomics because it avoids coupling the change to live session behavior.
- Mirroring the live workspace's checkbox style in the default scaffold makes future autoresearch workspaces easier to steer from the first run.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, or benign canary scoring on launch.

## Avoid Next Time

- Do not assume the default scaffold is “good enough” just because the current repo has a richer hand-edited planning file.
- Do not claim Hermes parity from better planning ergonomics or checklist scaffolding alone.

## Promising Next Directions

- Add a default scaffold section for blocked/non-goals or acceptance criteria if that can stay generic enough for non-Hermes autoresearch workspaces.
- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
