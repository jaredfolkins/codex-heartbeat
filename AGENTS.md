# AGENTS

## Purpose

`codex-heartbeat` is a Go wrapper around Codex that runs an autoresearch-style heartbeat loop:

- the human defines the goal in `program.md`
- the agent iterates in bounded cycles
- durable memory lives under `target/`

## Read These Files First

- `program.md`
  The current human-owned objective, evaluator, prompt mode, and constraints.
- `PLANNING.md`
  The live working plan. Keep only active tasks here.
- `target/PLANNING_HISTORY.md`
  Durable planning memory for completed, superseded, or restart-invalidated tasks.
- `target/latest-context.md`
  The compact summary that heartbeat prompt injections replay.
- `target/results.jsonl`
  The experiment ledger with hypothesis, command, outcome, disposition, and notes.
- `target/run-<timestamp>/{plan,execution,results,insights}.md`
  Per-run artifacts showing what the loop believed, did, observed, and learned.
- `README.md`
  User-facing contract and CLI behavior.

## File Ownership

- `program.md` is human-owned unless the user explicitly asks to change the objective or evaluator contract.
- `PLANNING.md` is agent-maintained and should stay concise and current.
- `target/PLANNING_HISTORY.md` is agent-maintained durable memory for planning history.
- `target/` is the bounded memory area; avoid scattering scratch notes elsewhere.

## Working Contract

- Read `program.md`, this file, `PLANNING.md`, and `target/latest-context.md` before acting.
- Keep each cycle bounded to one hypothesis and one primary evaluator.
- Update `target/run-<timestamp>/plan.md`, `execution.md`, `results.md`, and `insights.md`.
- Append concise result entries to `target/results.jsonl`.
- Move completed checklist items out of `PLANNING.md` into `target/PLANNING_HISTORY.md`.
- When a refactor or restart invalidates an old plan, preserve that trail in `target/PLANNING_HISTORY.md` instead of deleting it silently.

## Prompt Modes

- `autoresearch`
  Work the objective directly through the bounded experiment loop.
- `planning`
  Use the autoresearch loop to refine the goal, deepen the plan, and decide the next deep dive before broad implementation.
- `manual-test-first`
  Prepare the next candidate fix and exact validation steps, then stop before the final human gate.

## Validation

- If Go code changes, run focused tests first and `go test ./...` when practical.
- If screen-idle behavior changes, run the screen and replay tests.
- If prompt resolution, scaffolding, planning, or artifact behavior changes, run the autoresearch-focused tests.
- If validation cannot be run, say so explicitly.

## Template Sources

- Source templates now live under `cmd/codex-heartbeat/templates/`.
- Keep those Markdown templates aligned with the scaffold and prompt behavior in Go code.
