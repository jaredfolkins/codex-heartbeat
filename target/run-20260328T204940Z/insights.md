# Insights

## What Worked

- Challenging the “AGENTS is already complete” assumption still produced useful work: deeper layered tests instead of unnecessary runtime churn.
- Layered tests gave better failure localization than a monolithic replay harness would have.
- The existing runtime behavior held up under the stronger replay and ledger tests.

## What Failed

- Nothing new failed in runtime behavior; the gap was confidence, not functionality.

## Avoid Next Time

- Do not add broad replay frameworks before proving that smaller layered tests are insufficient.
- Do not change runtime heartbeat logic just because the checklist is complete; add evidence first.

## Promising Next Directions

- If future heartbeat regressions appear, expand the fixture corpus gradually rather than adding opaque runtime diagnostics.
- If live PTY behavior becomes a concern, capture a few sanitized real traces and keep them as a small holdout set.
