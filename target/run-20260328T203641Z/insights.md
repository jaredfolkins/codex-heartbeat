# Insights

## What Worked

- Reviewing the real upstream TUI snapshots immediately exposed which strings were status-row signals versus footer/history text.
- Keeping the scheduler untouched and hardening only the classifier made the change easy to validate.
- Recent-line weighting is a good fit for terminal snapshots because stale top-of-screen text no longer dominates the decision.

## What Failed

- The original phrase-only detector treated queued-message and background-terminal text as live work too often.

## Avoid Next Time

- Do not promote a phrase to a working-state signal unless it is tied to a current status-row shape.
- Do not rely on `esc to interrupt` alone; pending-steer UI reuses that wording.

## Promising Next Directions

- Capture a few real `screen-state.json` snapshots from live sessions and turn them into a fixture corpus.
- If future regressions appear, add lightweight classifier diagnostics so score reasoning is visible in logs.
