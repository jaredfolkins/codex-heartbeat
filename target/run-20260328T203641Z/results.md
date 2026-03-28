# Results

- Success: yes
- Primary evaluator: `go test ./...`
- Additional verification: `go test -race ./...`

## Observable Signals

- `go test ./cmd/codex-heartbeat -run 'Screen|TerminalScreen|EvaluateScreenIdlePoll'` passed.
- `go test ./...` passed.
- `go test -race ./...` passed.
- The detector now recognizes upstream-style active status rows such as `• Working (0s)`, `• Analyzing (0s • esc to interrupt)`, and wrapped `Waiting for background terminal` rows.
- Pending-steer previews, background-terminal footer summaries, and historical `Waited/Interacted with background terminal` lines no longer count as live work by themselves.

## Unexpected Behavior

- Upstream review showed that `Messages to be submitted after next tool call` belongs to pending-input preview UI, not a live running-state marker, so the old detector was more permissive than expected.
