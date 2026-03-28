# Insights

## What Worked

- Reducing an external "Godmode" concept into concrete primitives like system prompts, prefill, model-profile selection, and canary scoring made the task actionable.
- Grounding the checklist in both Hermes source and upstream Codex API surfaces avoided vague planning and exposed the real wrapper limitation quickly.
- Splitting the plan into tasks, non-goals, and acceptance criteria made the safe scope much harder to misread.
- Adding a prioritized phase-1 slice made the plan much more likely to turn into real implementation work on the next pass.
- Adopting the real upstream Codex CLI flags first was a low-risk way to make progress without pretending the wrapper already controls the full prompt stack.
- Passing reasoning effort as `--config model_reasoning_effort=...` kept the first implementation aligned with upstream Codex behavior.

## What Failed

- The X status link did not yield a reliable enough transcript to treat as a primary technical source.
- A direct "does the jailbreak work?" evaluator was not appropriate for this cycle.
- The wrapper still does not match Hermes's stronger feature set because the current transport cannot set base/developer instructions or prefill on launch.

## Avoid Next Time

- Do not treat "make GODMODE work" as a single prompt tweak when the source implementation actually depends on multiple launch-time instruction channels.
- Do not anchor a task plan on an X post alone when the linked repository already contains the concrete mechanics.
- Do not leave safety boundaries implicit when the feature name itself invites an unsafe interpretation.
- Do not claim Hermes parity from wrapper-safe model/profile pass-through alone.

## Promising Next Directions

- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
- Extend the new phase-1 launch metadata into user-facing docs so operators can verify which profile/model/effort was actually used.
