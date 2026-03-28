# AGENTS

Goal: make `codex-heartbeat` support an `autoresearch`-style autonomous loop instead of only re-injecting a generic standing prompt.

Status: completed in the current implementation pass on 2026-03-28.

## Prompt Model

- [x] Replace the embedded default prompt with an experiment-loop prompt that defines:
  - a single objective
  - a single primary evaluator
  - one hypothesis per cycle
  - keep/discard/revert semantics
  - explicit memory read/write behavior
- [x] Change the default behavior so the 3-agent council is a fallback for blocked or stalled states, not the first action on every idle heartbeat.
- [x] Add a first-class `program.md` convention so the human edits the research/debugging program while the agent edits the target workspace.
- [x] Define prompt precedence clearly across `--prompt`, repo-local `program.md`, and the embedded fallback prompt.
- [x] Add a prompt mode for "manual test first" workflows where the agent prepares the next candidate fix and validation steps, then stops before the final human gate.

## Memory And Artifacts

- [x] Add a reusable run-artifact layout under `target/run-<timestamp>/`.
- [x] Write `plan.md`, `execution.md`, `results.md`, and `insights.md` for each run.
- [x] Ingest prior `target/*/insights.md` before proposing the next step.
- [x] Maintain a compact experiment ledger such as `results.tsv` or `results.jsonl` with hypothesis, command, outcome, disposition, and notes.
- [x] Persist a short "latest context" summary so repeated heartbeat injections can remind the agent of recent findings without replaying the entire transcript.

## Execution Loop

- [x] Add an explicit experiment/debug loop contract:
  1. establish baseline
  2. pick one hypothesis
  3. make one bounded change
  4. run one evaluator command
  5. record result
  6. keep or discard
- [x] Support evaluator commands that can be recorded and reused, for example `make manual-lab-up`.
- [x] Detect repeated failures or dead ends and trigger the council only after a configurable threshold.
- [x] Add optional git checkpoint support for "save point" commits after meaningful progress.
- [x] Keep notes and artifacts tidy by default so autonomous runs do not scatter files across the workspace.

## UX And Scaffolding

- [x] Add a scaffold/init command for an `autoresearch`-style workspace that creates `program.md`, a results ledger, and run-artifact templates.
- [x] Ship example prompts for:
  - debugging loops
  - benchmark/experiment loops
  - manual-validation loops
- [x] Update the README to explain the `autoresearch` mental model:
  - human edits the program
  - agent runs the loop
  - artifacts hold memory
  - evaluator decides progress
- [x] Document a recommended repo contract for target workspaces, including `AGENTS.md`, `PLANNING.md`, and `target/*/insights.md`.

## Tests

- [x] Add tests for prompt source precedence and fallback behavior.
- [x] Add tests for artifact discovery and prior-insight ingestion.
- [x] Add tests for any scaffold/init command output.
- [x] Add tests for evaluator-command recording and result-ledger writes.
- [x] Add tests for council-trigger thresholds so the fallback only appears when the loop is genuinely stuck.
