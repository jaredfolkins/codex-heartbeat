# Plan

- Hypothesis: A recent-line evidence classifier grounded in real upstream Codex status rows will reduce false heartbeat fires without regressing idle detection.

## Steps

1. Ingest prior `target/*/insights.md` artifacts and summarize what to reuse.
2. Clone upstream `openai/codex` into an ignored workspace-local directory and review the TUI status rendering code and snapshots.
3. Run a 3-agent council, critique the candidate detector approaches, and select the most feasible plan.
4. Replace broad phrase matching in `screen.go` with recent-line evidence scoring plus strong negative guards.
5. Add upstream-derived tests for hidden interrupt hints, wrapped status rows, pending-steer previews, history/footer text, and stale-screen text.
6. Validate with focused tests, `go test ./...`, and `go test -race ./...`.

## Assumptions

- `codex-heartbeat` should keep its current scheduler contract and improve classification quality first.
- The best proxy for live work is a recent status-row shape, not any historical phrase visible anywhere on screen.
- The cloned upstream repo must stay ignored and out of the save-point commit.
