# Planning

## Objective

- Keep the current objective aligned with program.md.

## Evaluator

- Reuse one primary evaluator until the hypothesis changes.

## Open Questions

- Capture blockers or questions the next loop should resolve.

## GODMODE-Style Prompt Profile Feasibility

### Review Basis

- X post: `https://x.com/KamakuraCrypto/status/2037294903814738261?s=20`
  It frames the target workflow as a multi-LLM research / cross-review pattern rather than a single static prompt.
- Hermes repo: `https://github.com/nousresearch/hermes-agent`
  It shows that Hermes combines stronger launch-time instruction control, ephemeral prefill, canary-style evaluation, and delegated parallel review surfaces.

### Source Review Notes

- Hermes's `godmode` feature is not a single prompt. It combines model-family strategy selection, launch-time system prompt injection, ephemeral prefill messages, and canary scoring to decide whether a profile "worked".
- Hermes also exposes named personalities and project context files as first-class UX surfaces, so reusable instruction bundles are operator-visible concepts instead of only hidden transport settings.
- Hermes pairs personality-style changes with obvious fresh-session flows (`/new` or `/reset`), so operators have a clear way to start a conversation under the newly selected bundle.
- Hermes spans long-lived conversations and cross-interface use, so the operator model also needs a clear persistence rule for whether a selected bundle is session-local, repo-default, or sticky across future sessions.
- The current `codex-heartbeat` wrapper mostly injects user-visible prompts into an existing Codex thread. Its interactive path currently launches `codex` or `codex resume` without first-class `base_instructions` or `developer_instructions` overrides.
- Upstream Codex app-server and SDK surfaces do expose `base_instructions`, `developer_instructions`, and `config.model_reasoning_effort`, so a stronger prompt-stack feature likely requires an app-server or SDK-backed path rather than more user-message reinjection.

### Task List

- [ ] Define the safe scope for this feature: prompt-stack and instruction-profile testing only, not safety-bypass or jailbreak claims.
- [ ] Add a first-class prompt profile concept for `codex-heartbeat` such as `--profile`, a repo-local profile file, or a `program.md` stanza that can describe model, reasoning effort, base instructions, developer instructions, and optional prefill.
- [ ] Make prompt profiles operator-visible like Hermes personalities/context files, so reusable instruction bundles can be selected by name instead of only assembled from low-level flags.
- [ ] Add a discoverable profile-selection surface, such as status/help output or an interactive command, so operators can list and switch named bundles without editing files blindly.
- [ ] Define profile-switch scope semantics explicitly, so operators know whether changing a named bundle affects the current conversation, the next fresh thread, or only newly created sessions.
- [ ] If profile switching is next-session-only, add a one-step reset/new-session flow so operators can start a fresh conversation under the selected bundle without manual file edits or ambiguous restart steps.
- [ ] If profile switching is deferred, show both the active and pending bundle in wrapper UX so operators can tell what is in effect now versus what will apply after reset/new session.
- [ ] Define profile persistence scope explicitly, so operators know whether selecting a bundle changes only the current session, the repo-local default, or future sessions started from the same workspace.
- [ ] Decide the transport boundary for the feature: keep the current CLI-wrapper path for heartbeat reinjection, or add a Codex SDK/app-server backend for sessions that need true `base_instructions` / `developer_instructions`.
- [ ] Add launch-time instruction injection for both new threads and resumed threads, because the current `buildInteractiveArgs()` path only starts `codex` or `codex resume` and cannot set upstream instruction fields.
- [ ] Add optional ephemeral prefill support so the wrapper can seed the first turn or thread history without writing persistent prompt hacks into workspace files by default.
- [ ] Add model-profile selection logic so `gpt-5.3-codex-spark` with `high` reasoning can use a named instruction profile instead of the generic heartbeat prompt.
- [ ] Add a Hermes-style delegated cross-review surface for benign evaluator and council work so multi-agent review is a first-class workflow instead of an improvised fallback.
- [ ] Add a safe evaluator harness that tests harmless instruction-following canaries on `gpt-5.3-codex-spark` with `high` reasoning, so we can verify whether the prompt stack sticks without attempting to bypass safeguards.
- [ ] Record the selected profile, model, reasoning effort, and instruction-source metadata in `target/` artifacts and runtime logs for reproducibility.
- [ ] Record source-traceability metadata such as `review_basis` anywhere parity claims are surfaced so operators can see which Hermes materials the comparison is grounded in.
- [ ] Add tests for profile precedence, new-vs-resume session behavior, evaluator recording, and any SDK/app-server integration seam.
- [ ] Document the limitation clearly: user-message heartbeat injections are weaker than base/developer instruction channels, so "GODMODE works" is not a meaningful claim unless the wrapper controls the full prompt stack.

### Blocked / Non-Goals

- [ ] Do not implement or advertise a jailbreak, safety-bypass, or refusal-suppression feature.
- [ ] Do not use harmful or disallowed prompts as the success criterion for this work.
- [ ] Do not claim parity with Hermes `godmode` unless `codex-heartbeat` can actually control the same launch-time instruction channels and has a benign evaluator harness to prove it.

### Acceptance Criteria For The Safe Alternative

- [ ] A user can select a named prompt profile for `gpt-5.3-codex-spark` with `high` reasoning without editing the wrapper source.
- [ ] A user can select a named profile/personality-style bundle or repo-local context-file equivalent that predictably shapes new conversations.
- [ ] A user can discover the available named bundles and the current selection from the wrapper's own UX instead of inferring it from file contents.
- [ ] A user can tell whether switching profiles takes effect immediately for the active session or only for the next conversation, and the wrapper behaves consistently with that rule.
- [ ] If immediate in-thread switching is not supported, a user can start a fresh session with the selected bundle in one obvious wrapper-supported step.
- [ ] If a profile change is pending, the wrapper clearly shows both the active bundle and the queued next-session bundle.
- [ ] A user can tell whether a selected bundle is a one-session override or a persisted default for future sessions in the same workspace.
- [ ] The chosen profile can control model selection, reasoning effort, and at least one stronger instruction channel than a plain user-message heartbeat.
- [ ] New and resumed sessions behave predictably, and any profile override is visible in runtime logs and `target/` artifacts.
- [ ] A harmless evaluator can verify that the selected profile changes instruction-following behavior in a measurable, repeatable way.
- [ ] The parity explanation stays traceable to reviewed source material instead of only local shorthand.
- [ ] The docs clearly separate "prompt-profile support" from any unsupported or unsafe "GODMODE" expectation.

### Phase 1 Recommendation

- [ ] Start with a repo-local prompt profile file and `--profile` flag that selects `model`, `model_reasoning_effort`, and a named instruction bundle.
- [ ] Keep the phase-1 UX close to Hermes's personality/context-file model so the selected instruction bundle is a visible operator choice, not just hidden launch plumbing.
- [ ] Make the selected profile discoverable in the same phase through `status`, help text, or an equivalent UX so the operator can confirm which bundle is active.
- [ ] Decide in the same phase whether profile switching is "apply to current session", "apply on next reset/new session", or both, and surface that rule in the operator UX.
- [ ] If phase 1 uses next-session-only switching, pair it with a reset/new-session command or workflow so changing bundles does not require manual wrapper restarts.
- [ ] If phase 1 uses deferred switching, expose an active-versus-pending profile state in `status`, help, or equivalent UX so operators can confirm what will happen on reset/new session.
- [ ] Decide in the same phase whether bundle selection persists across future sessions or only applies as a one-session override, and surface that rule in operator UX and artifacts.
- [ ] Keep the first implementation on the current wrapper path, but limit the scope to fields the wrapper can already pass safely; treat any need for true `base_instructions` / `developer_instructions` as the trigger for a later SDK/app-server phase.
- [ ] Add logging and `target/` artifact capture for the selected profile name, model, and reasoning effort in the same patch so validation stays observable.
- [ ] Carry `review_basis` or equivalent source-traceability evidence through the same phase-1 status/help/docs surfaces so parity claims stay auditable while the transport is still wrapper-based.
- [ ] Decide whether the existing council path should grow into a first-class delegated cross-review surface before attempting any transport-layer refactor.
- [ ] Add a benign canary evaluator for the selected profile before attempting any transport-layer refactor.

### Hermes Parity Gap

- [ ] Add a stronger launch-time instruction channel than plain user-message reinjection, because Hermes `godmode` depends on system-style instruction control.
- [ ] Add optional ephemeral prefill for new and resumed sessions so the wrapper can shape the first turn without persisting prompt hacks into workspace files.
- [ ] Add a harmless canary-scoring harness that can distinguish "profile attached" from "profile actually changed behavior" in a repeatable way.
- [ ] Add a Hermes-style delegated cross-review workflow so the wrapper can support the reviewed multi-LLM research pattern, not just single-agent prompt stacking.
- [ ] Keep parity claims source-grounded with explicit review-basis evidence instead of relying on local wording alone.
- [ ] Define a parity claim rule that stays false until the wrapper can prove equivalent launch-time control and benign evaluation coverage.
