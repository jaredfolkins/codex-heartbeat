# Latest Context

- Objective: Update the docs so heartbeat behavior is easier to triage and verify from stable artifacts and focused test commands.
- Primary evaluator: `go test ./...`
- Prompt mode: `documentation + verification`
- Recent failure streak: 0 / 3

## Recent Ledger
- `keep` via `go test ./...`: pass | Added a README playbook for heartbeat triage and contributor verification, validated the documented focused test commands, and kept the docs artifact-first instead of implementation-frozen.
- `keep` via `go test ./...`: pass | Ran a 3-agent council against the already-complete checklist, chose layered classifier/scheduler/council-threshold tests, added screen fixtures plus replay and ledger robustness tests, and passed focused, full, and race suites.
- `keep` via `go test ./...`: pass | Cloned upstream Codex into ignored tmp/openai-codex, reviewed live status-row snapshots, replaced phrase-only matching with weighted recent-line evidence, and passed focused, full, and race tests.
- `keep` via `go test ./...`: pass | Implemented program.md precedence, bounded target/ memory artifacts, init scaffolding, examples, README updates, and passing unit+races tests.

## Prior Insights
- run-20260328T204940Z/insights.md: Challenging the “AGENTS is already complete” assumption still produced useful work: deeper layered tests instead of unnecessary runtime churn. Layered tests gave better failure localization than a monolithic replay harness would have. The existing runtime behavior held up under the stronger replay and ledger tests. Nothing new failed in runtime behavior; the gap was confidence, not functionality.
- run-20260328T203641Z/insights.md: Reviewing the real upstream TUI snapshots immediately exposed which strings were status-row signals versus footer/history text. Keeping the scheduler untouched and hardening only the classifier made the change easy to validate. Recent-line weighting is a good fit for terminal snapshots because stale top-of-screen text no longer dominates the decision. The original phrase-only detector treated queued-message and background-terminal text as live work too often.
- run-20260328T201941Z/insights.md: Splitting the work into prompt-resolution first and artifact/memory wiring second kept the refactor manageable. A dedicated `autoresearch.go` file kept the new behavior out of the hottest CLI paths. The results ledger is a practical place to drive council fallback thresholds without changing scheduler semantics. The first integration pass left one deterministic screen-log filename bug, which the test suite caught immediately.
