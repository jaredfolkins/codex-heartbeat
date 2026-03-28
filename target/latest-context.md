# Latest Context

- Objective: Improve the screen-idle heartbeat detector against real upstream Codex TUI status patterns.
- Primary evaluator: `go test ./...`
- Prompt mode: `manual review + detector hardening`
- Recent failure streak: 0 / 3

## Recent Ledger

- `keep` via `go test ./...`: pass | Replaced phrase-only screen classification with recent-line evidence scoring, added upstream-derived negative guards, and passed unit + race tests.

## Prior Insights

- Stage refactors so scheduler semantics stay stable while detector logic changes.
- Avoid broad phrase matches that treat queued-message or history/footer text as live work.
