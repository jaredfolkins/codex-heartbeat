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
- Hermes also documents persistent memory and user profiles, so the operator model needs a clear rule for whether prompt-profile selection interacts with saved memory or remains an isolated instruction-layer choice.
- Hermes also separates global identity/personality concerns from project context files, so the operator model needs an explicit precedence rule for which layer wins when a named bundle, global persona, and repo-local context all provide overlapping instructions.
- Hermes also exposes a distinct global identity file contract: `SOUL.md` lives only in `HERMES_HOME`, occupies identity slot `#1`, stays separate from project context discovery, and falls back to a built-in default identity when absent or empty.
- Hermes treats `AGENTS.md` as a hierarchical project-context input, combining multiple files across a monorepo instead of assuming a single repo-root instruction file.
- Hermes's hierarchical project context is also ordered and labeled: discovered files are concatenated with relative path headers, so operators can tell where each repo-local instruction block came from.
- Hermes also recognizes adjacent instruction-file conventions like `.cursorrules`, so the operator model needs a compatibility rule for whether existing repo guidance files are ignored, imported, or merged alongside `AGENTS.md`.
- Hermes also exposes background sessions as a first-class operator workflow: parallel tasks get isolated conversation history while inheriting the current session's model and reasoning-style settings.
- Hermes also supports launch-time skill and toolset loading, so the operator model needs a rule for whether a named bundle can preload extra tools/skills in addition to prompt instructions.
- Hermes's background-session behavior inherits more than just model choice: the child session also carries provider, toolsets, reasoning settings, and fallback-model configuration from the parent session.
- Hermes also exposes isolated git worktrees for parallel agents, so the operator model needs an explicit rule for whether delegated review tasks should run in conflict-avoiding workspaces instead of sharing one mutable checkout.
- Hermes also makes background work operator-visible: each task gets a numbered label and standalone task ID, and completion/error is delivered on a separate result surface instead of silently mutating the main conversation history.
- Hermes's delegated subagents also have an explicit parent-return boundary: children run with fresh context and restricted toolsets, and only a final structured summary re-enters the parent context instead of the full child transcript.
- Hermes's delegated subagents also use explicit toolset narrowing, so the operator model needs a rule for whether child agents may be limited to a subset of the parent tool surface and how that restriction is made visible.
- Hermes's delegated subagents also have an explicit recursion boundary: delegation is depth-limited, with child agents unable to spawn grandchildren, so parallel review cannot fan out indefinitely.
- Hermes's delegated subagents also have an explicit context-handoff contract: children start with zero knowledge of the parent session and receive only the packaged goal/context the parent provides.
- Hermes's delegated subagents also return a structured summary shape, covering what they did, what they found, files modified, and issues encountered rather than an unstructured child response blob.
- Hermes's delegated subagents also have an explicit concurrency cap: parallel batches run only up to a bounded number of concurrent child agents instead of allowing unlimited fan-out in a single parent turn.
- Hermes also exposes configurable background-notification behavior: operators can choose whether long-running background work pushes all updates, result-only updates, error-only updates, or no updates, and CLI completion can optionally ring a terminal bell.
- Hermes also exposes explicit interrupt-and-redirect behavior: sending a new message while work is in flight interrupts the run, kills active terminal commands, cancels queued tool calls, combines interruption messages into one prompt, and supports stopping without queuing a redirect.
- Hermes's delegated subagents also have explicit interrupt propagation: interrupting the parent interrupts all active children instead of leaving background child work detached from the parent's control flow.
- Hermes's delegated subagents also have explicit progress-display behavior: CLI mode shows a real-time tree-view of child tool calls with per-task completion lines, while gateway mode batches delegated progress back through the parent progress callback.
- Hermes also exposes configurable tool-activity display behavior: operators can choose whether tool progress is off, only new tool starts, all tool activity, or verbose output, and messaging can optionally expose a `/verbose` toggle for that surface.
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
- [ ] Add an explicit return-to-default flow so operators can clear a named bundle and get back to the repo or wrapper default without manual file edits.
- [ ] Define profile persistence scope explicitly, so operators know whether selecting a bundle changes only the current session, the repo-local default, or future sessions started from the same workspace.
- [ ] Record profile-change history in operator-visible artifacts or status so it is clear when a bundle was selected, cleared, deferred, or promoted to the default.
- [ ] Define how prompt-profile selection interacts with persistent memory or user-profile state, so operators know whether choosing a bundle changes only instructions or also affects saved context.
- [ ] Define instruction-layer precedence explicitly, so operators know whether global persona/profile settings or repo-local context files win when they overlap.
- [ ] Define global-identity file semantics explicitly, so operators know whether a per-user identity file exists outside repo context, where it is loaded from, and what fallback identity applies when that file is absent or empty.
- [ ] Define repo-context discovery semantics explicitly, so operators know whether project instructions come from one repo-local file or a hierarchical set of `AGENTS.md` files.
- [ ] Define repo-context merge semantics explicitly, so operators know the order in which hierarchical `AGENTS.md` files are combined and how each instruction block is labeled or traced back to its source path.
- [ ] Define context-file compatibility semantics explicitly, so operators know whether existing repo guidance files like `.cursorrules` are ignored, imported, or merged alongside `AGENTS.md`.
- [ ] Define background/delegated-session inheritance semantics explicitly, so operators know whether parallel review tasks inherit the active profile/model/reasoning settings while keeping isolated conversation history.
- [ ] Define bundle-to-tooling semantics explicitly, so operators know whether selecting a named bundle also preloads toolsets or skills in addition to prompt/context instructions.
- [ ] Define delegated-session configuration inheritance explicitly, so operators know whether provider, toolsets, reasoning settings, and fallback model carry into background review tasks alongside the active profile.
- [ ] Define delegated-workspace semantics explicitly, so operators know whether parallel review tasks run in isolated git worktrees or share the primary checkout.
- [ ] Define delegated-task result-surfacing semantics explicitly, so operators know how background review tasks are identified, tracked, and delivered back to the main workflow instead of disappearing into hidden worker state.
- [ ] Define delegated-summary boundary semantics explicitly, so operators know whether child-agent work returns only a final summary to the parent context or merges richer child transcript and tool-trace state.
- [ ] Define delegated-toolset restriction semantics explicitly, so operators know whether child agents may be limited to a subset of tools and how that narrowed tool surface is selected and surfaced.
- [ ] Define delegated-depth-limit semantics explicitly, so operators know whether child agents may delegate further or whether delegation stops at one child layer with no grandchildren.
- [ ] Define delegated-context handoff semantics explicitly, so operators know what parent context is packaged for a child agent instead of assuming the child inherits ambient conversation state.
- [ ] Define delegated-summary schema semantics explicitly, so operators know what a child agent must report back, including findings, files touched, and issues encountered, instead of accepting an unstructured return.
- [ ] Define delegated-concurrency-cap semantics explicitly, so operators know how many child agents may run in parallel for one parent task instead of assuming unlimited delegated fan-out.
- [ ] Define background-notification policy semantics explicitly, so operators know whether long-running background work sends all updates, result-only updates, error-only updates, or no updates, and whether completion can ring a bell.
- [ ] Define interrupt-and-redirect semantics explicitly, so operators know what happens when they send follow-up input during in-flight work, including command termination, queued-tool cancellation, message coalescing, and stop-without-redirect behavior.
- [ ] Define delegated-interrupt-propagation semantics explicitly, so operators know whether interrupting the parent also interrupts all active child agents or leaves delegated work running independently.
- [ ] Define delegated-progress-display semantics explicitly, so operators know whether child progress is shown as real-time per-task tool activity, batched parent progress, or some other surface instead of leaving delegated progress visibility implicit.
- [ ] Define tool-activity display semantics explicitly, so operators know whether tool progress is off, new-only, all, or verbose, and whether there is an operator-visible toggle for changing that surface.
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
- [ ] A user can intentionally clear a selected bundle and confirm that the wrapper has returned to the default profile state.
- [ ] A user can tell whether a selected bundle is a one-session override or a persisted default for future sessions in the same workspace.
- [ ] A user can inspect a timestamped record of recent profile changes so active and pending bundle state is auditable instead of purely ephemeral.
- [ ] A user can tell whether a selected bundle affects persistent memory/user-profile context or only the prompt stack for the current workspace/session.
- [ ] A user can tell which instruction layer wins when global persona settings, a named bundle, and repo-local context files overlap.
- [ ] A user can tell whether a global identity file exists outside repo context, whether it loads only from a global home path, whether it occupies a dedicated top-level identity slot, and what fallback applies when it is missing or empty.
- [ ] A user can tell whether repo context comes from one file or a hierarchical set of `AGENTS.md` files and can explain which files were loaded.
- [ ] A user can tell the order in which hierarchical repo-context files were merged and can trace each injected block back to a relative path header or equivalent source label.
- [ ] A user can tell whether existing repo instruction files like `.cursorrules` participate in the assembled context and how they were mapped into the wrapper's repo-context model.
- [ ] A user can tell whether a spawned background/delegated task inherits the active profile/model/reasoning settings and whether its conversation history is isolated from the parent session.
- [ ] A user can tell whether a selected named bundle also loads extra toolsets or skills and can inspect that tooling surface from wrapper UX or artifacts.
- [ ] A user can tell whether a background/delegated task also inherits provider, toolsets, and fallback-model settings from the parent session instead of only model/reasoning.
- [ ] A user can tell whether a background/delegated task runs in an isolated worktree and can trace which workspace produced its artifacts or code changes.
- [ ] A user can tell how a background/delegated task is identified, where its result appears, and whether completion/error delivery is separate from the main conversation history.
- [ ] A user can tell whether delegated child-agent work re-enters the parent only as a final summary or whether richer child transcripts and tool traces are exposed separately.
- [ ] A user can tell whether a delegated child agent is restricted to a subset of tools and can inspect which toolsets were allowed for that child instead of inferring it from implementation details.
- [ ] A user can tell whether delegated child agents may themselves spawn more children or whether delegation is intentionally capped at one child layer with no recursive fan-out.
- [ ] A user can tell what goal/context package a delegated child actually receives and can confirm that the child does not silently inherit ambient parent conversation state.
- [ ] A user can tell what summary fields a delegated child must return, including what it did, what it found, files modified, and unresolved issues, instead of receiving an opaque child response.
- [ ] A user can tell how many delegated child agents may run concurrently for one parent task and whether excess delegated work is queued, rejected, or serialized.
- [ ] A user can tell what background-notification policy is active for long-running delegated/background work, including whether updates are `all`, `result`, `error`, or `off`, and whether completion rings a terminal bell.
- [ ] A user can tell whether interrupting in-flight work kills running terminal commands, cancels queued tool calls, combines multiple interruption messages, and supports stopping without queuing a new redirect prompt.
- [ ] A user can tell whether interrupting the parent also interrupts all active delegated children and whether any delegated work may outlive the parent interruption boundary.
- [ ] A user can tell whether delegated child progress appears as a real-time CLI tree-view with per-task completion lines, as batched parent progress updates, or through some other declared progress surface.
- [ ] A user can tell whether tool activity is hidden, new-only, all, or verbose, and whether an operator-visible toggle can change that display surface without editing implementation internals.
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
- [ ] If phase 1 adds profile selection, include a matching clear/default action so the operator can undo an override without editing config by hand.
- [ ] Decide in the same phase whether bundle selection persists across future sessions or only applies as a one-session override, and surface that rule in operator UX and artifacts.
- [ ] If phase 1 adds profile selection, record profile-change events in `target/` artifacts or `status` so deferred and persisted changes remain explainable after the fact.
- [ ] In phase 1, either keep profile selection isolated from persistent memory/user-profile state or document the exact interaction explicitly in status/help/artifacts.
- [ ] In phase 1, document a simple precedence rule across global persona settings, named bundles, and repo-local context files so overlapping instructions are explainable before any richer transport work starts.
- [ ] In phase 1, document whether a global identity file exists outside repo context, whether it loads only from a home-directory path, whether it occupies a dedicated identity slot, and what built-in fallback applies when it is missing or empty.
- [ ] In phase 1, document whether repo context uses a single-file contract or hierarchical `AGENTS.md` discovery so monorepo behavior is predictable before transport work starts.
- [ ] In phase 1, if repo context is hierarchical, document the merge order and expose source labels or relative path headers so injected instructions remain explainable in monorepos.
- [ ] In phase 1, document whether compatibility files like `.cursorrules` are supported and, if so, how they map into the repo-context assembly and observability surfaces.
- [ ] In phase 1, if background/delegated sessions are exposed, document whether they inherit the active profile/model/reasoning settings and whether their histories remain isolated from the parent session.
- [ ] In phase 1, document whether named bundles may preload toolsets or skills and, if so, surface that loaded tooling set in status/help/artifacts.
- [ ] In phase 1, if delegated/background sessions are exposed, document whether provider, toolsets, and fallback model inherit from the parent session and show that inherited execution config in status/help/artifacts.
- [ ] In phase 1, if delegated/background review is exposed, document whether workers run in isolated git worktrees and surface that workspace choice in status/help/artifacts.
- [ ] In phase 1, if delegated/background review is exposed, document how task IDs, progress, and result delivery appear in status/help/artifacts so delegated work stays auditable.
- [ ] In phase 1, if delegated child-agent review is exposed, document whether only a final summary re-enters the parent context or whether richer child transcripts/artifacts are surfaced separately, and keep that boundary visible.
- [ ] In phase 1, if delegated child-agent review is exposed, document whether child tool access is narrowed to a subset of tools and surface the allowed toolsets in status/help/artifacts so delegation scope is explainable.
- [ ] In phase 1, if delegated child-agent review is exposed, document whether child agents may delegate further or whether recursion is capped at one child layer, and keep that limit visible in status/help/artifacts.
- [ ] In phase 1, if delegated child-agent review is exposed, document what goal/context handoff the parent supplies to the child and surface that packaged context boundary in status/help/artifacts so delegation inputs stay explainable.
- [ ] In phase 1, if delegated child-agent review is exposed, document the expected child-summary fields and surface that schema in status/help/artifacts so delegated outputs stay comparable and auditable.
- [ ] In phase 1, if delegated child-agent review is exposed, document the maximum concurrent child-agent fan-out and surface that cap in status/help/artifacts so delegated parallelism stays predictable.
- [ ] In phase 1, if delegated/background work is exposed, document the background-notification policy and surface whether updates are `all`, `result`, `error`, or `off`, plus whether completion can ring a bell.
- [ ] In phase 1, if interactive interruption is exposed, document the interrupt-and-redirect policy and surface whether running terminal commands are killed, queued tools are cancelled, interruption messages are combined, and stop-without-follow-up is supported.
- [ ] In phase 1, if delegated child-agent review is exposed, document whether parent interruption propagates to all active children and surface that cancellation boundary in status/help/artifacts so delegated control flow stays predictable.
- [ ] In phase 1, if delegated child-agent review is exposed, document whether progress appears as a real-time CLI tree-view, batched parent progress callback updates, or another declared surface so delegated progress stays predictable across interfaces.
- [ ] In phase 1, if tool activity is surfaced, document whether display is `off`, `new`, `all`, or `verbose`, and whether an operator-visible toggle like `/verbose` exists so tool visibility stays predictable across interfaces.
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
