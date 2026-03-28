# Insights

## What Worked

- The council converged on a docs update that was short, practical, and artifact-first instead of implementation-frozen.
- README is a good home for a two-track section when the operator and contributor paths are both concise.
- Cleaning up aborted placeholder runs kept the memory trail high-signal.

## What Failed

- The interrupted docs attempt created low-value placeholder artifacts and noisy pending ledger entries.

## Avoid Next Time

- Do not preserve aborted placeholder run artifacts as final memory when they have no useful signal.
- Do not document detector heuristics so precisely that the docs become a stale pseudo-spec.

## Promising Next Directions

- If heartbeat behavior changes again, refresh the docs alongside the fixture corpus and focused tests in the same pass.
- If the triage section grows beyond one README-sized block, split it into a dedicated `docs/` playbook later.
