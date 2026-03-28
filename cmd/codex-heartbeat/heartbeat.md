Autoresearch loop contract

Objective: {{OBJECTIVE}}
Primary evaluator: {{PRIMARY_EVALUATOR}}
Prompt mode: {{PROMPT_MODE}}
Council after failures: {{COUNCIL_AFTER_FAILURES}}
Current run directory: {{RUN_DIR}}
Latest context path: {{LATEST_CONTEXT_PATH}}

Read memory before acting:
- Read `{{LATEST_CONTEXT_PATH}}` before you choose the next step.
- Reuse prior `target/*/insights.md` and `target/results.jsonl` only as bounded memory, not as an excuse to restart from scratch.
- Treat the human-authored program below as the authoritative research/debugging program when it exists.

Write memory during this cycle:
- Update `{{RUN_DIR}}/plan.md`, `{{RUN_DIR}}/execution.md`, `{{RUN_DIR}}/results.md`, and `{{RUN_DIR}}/insights.md`.
- Append one concise line to `target/results.jsonl` with hypothesis, command, outcome, disposition, and notes.
- Keep `insights.md` concise and high-signal.

Loop contract:
1. Establish the current baseline.
2. Pick exactly one hypothesis for this cycle.
3. Make one bounded change.
4. Run exactly one primary evaluator.
5. Record the observable result.
6. Decide keep, discard, or revert before moving on.

Council policy:
{{COUNCIL_INSTRUCTION}}

Checkpoint policy:
{{CHECKPOINT_INSTRUCTION}}

Manual gate policy:
{{MANUAL_GATE_INSTRUCTION}}

Human program:
{{PROGRAM_BODY}}

Latest context summary:
{{LATEST_CONTEXT}}
