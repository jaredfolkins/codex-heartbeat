# Insights

## What Worked

- `PLANNING.md` is still the best place to express the requested `[ ]` implementation backlog once the status/help/docs surfaces already agree.
- A simple grep evaluator was enough to keep the planning cycle bounded while still checking that the exact reviewed links showed up in the planning artifact.
- The two reviewed links belong in the planning artifact itself, not only in run memory or implicit source notes.
- Hermes's operator-visible personality/context-file UX is worth tracking explicitly in the backlog; otherwise the plan over-focuses on hidden prompt plumbing and under-specifies the user-facing profile model.
- Hermes-style named bundles also need a discoverability/switching story; a backlog that only defines profiles in files still misses an operator-facing part of the reviewed workflow.
- Named bundles also need explicit session-scope rules; without that, operators still cannot tell whether a switch should affect the current thread or only future sessions.
- If switching only applies to future sessions, the backlog also needs an obvious reset/new-session path; otherwise the UX remains underspecified even after the scope rule is written down.
- If switching is deferred, the backlog also needs an active-versus-pending state model; otherwise operators still cannot tell what profile is live before they reset the session.
- Named bundles also need a persistence rule; without it, operators still cannot tell whether a selection is a temporary override or the new default for later sessions.

## What Failed

- The wrapper still does not match Hermes's stronger feature set because it still cannot set base/developer instructions, prefill, benign canary scoring, or Hermes-style delegated cross-review on launch.

## Avoid Next Time

- Do not let the planning backlog drift behind the already-documented parity and traceability surfaces.
- Do not claim Hermes parity from a more complete task list alone.

## Promising Next Directions

- Prototype a non-destructive SDK/app-server-backed mode that can set `base_instructions`, `developer_instructions`, model, and reasoning effort for new and resumed Codex threads.
- Decide whether the existing fallback council needs a more Hermes-like first-class delegated review surface for benign evaluator comparisons.
- Add a harmless prompt-adherence harness for `gpt-5.3-codex-spark` with `high` reasoning so profile effectiveness can be measured without trying to bypass safeguards.
- Decide whether the first safe profile implementation should surface named bundles via `program.md`, a repo-local profile file, or both so it better matches Hermes's personality/context-file UX.
- Decide whether phase 1 should expose profile listing/switching through `status`, help, or an interactive command so the active bundle is obvious at runtime.
- Decide whether phase 1 should make profile switches immediate, next-session-only, or dual-mode, and log that behavior clearly when a user changes bundles.
- If phase 1 lands on next-session-only switching, design the exact reset/new-session command flow early so profile switching does not feel half-finished.
- If phase 1 lands on deferred switching, decide exactly where active and pending bundle state should appear so status/help UX stays unambiguous.
- If phase 1 allows persisted bundle selection, decide where that default lives and how a one-session override differs in status/help/artifacts.
