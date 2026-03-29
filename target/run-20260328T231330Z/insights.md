# Insights

## What Worked

- A small `status` payload change was enough to make the current non-equivalence answer better match the reviewed Hermes references without changing runtime behavior.
- Reusing the focused status tests kept the cycle bounded while still validating both the parity surface and the launch-settings surface together.
- The Hermes README plus the X post fragment were enough to justify naming delegated cross-review as a first-class parity gap instead of leaving it implicit.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, benign canary scoring, or Hermes-style delegated cross-review on launch.

## Avoid Next Time

- Do not rely only on local shorthand when the user supplied upstream references that clarify which capabilities Hermes is actually known for.
- Do not claim Hermes parity from a more complete task list alone.

## Promising Next Directions

- Consider whether README should mention the delegated cross-review gap now that `status` exposes it directly.
- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Decide whether the existing fallback council needs a more Hermes-like first-class delegated review surface for benign evaluator comparisons.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
