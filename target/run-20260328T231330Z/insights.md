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
- Named bundles also need a clear undo path; otherwise the plan describes how to apply overrides more clearly than how to remove them.
- Named bundles also need change history; otherwise status may show the current state but not how or when it got there.
- Named bundles also need an explicit memory interaction rule; otherwise it stays unclear whether selecting a bundle changes only prompt instructions or also saved context.
- Named bundles also need an explicit precedence rule against global persona settings and repo-local context files; otherwise overlapping instructions remain ambiguous even when all three layers are visible.
- Repo context also needs an explicit discovery rule; otherwise a safe implementation may assume one repo-local file where Hermes actually models a hierarchy of `AGENTS.md` files.
- Hierarchical repo context also needs explicit merge-order and source-label rules; otherwise discovering multiple files still leaves operators guessing how those instructions were assembled.
- Repo context also needs a compatibility rule for adjacent instruction-file conventions; otherwise the plan can model Hermes's hierarchy but still miss existing repo guidance that lives outside `AGENTS.md`.
- Delegated/background review also needs explicit inheritance and isolation rules; otherwise a parallel-workflow design can exist on paper while leaving it unclear which active settings carry over to spawned tasks.
- Named bundles also need explicit tooling semantics; otherwise a plan can match Hermes's prompt surfaces while still missing that launch-time bundles may carry toolsets or skills too.
- Delegated/background review also needs explicit full-config inheritance rules; otherwise a plan can say tasks inherit the active profile while still leaving provider, toolsets, and fallback-model behavior underspecified.
- Delegated/background review also needs explicit workspace-isolation rules; otherwise a multi-agent workflow can sound safe on paper while still leaving file-collision behavior unspecified.
- Delegated/background review also needs explicit task/result surfacing rules; otherwise operators still cannot tell how spawned work is identified or where completion/error is delivered.
- Delegated child-agent review also needs an explicit summary-return boundary; otherwise operators still cannot tell whether only the child summary or the full child transcript/tool state flows back into the parent context.
- Delegated child-agent review also needs an explicit tool-narrowing rule; otherwise operators still cannot tell whether a child inherits the parent tool surface or runs under a smaller allowed toolset.
- Delegated child-agent review also needs an explicit recursion limit; otherwise operators still cannot tell whether child agents may spawn grandchildren or whether delegation intentionally stops after one layer.
- Delegated child-agent review also needs an explicit context-handoff rule; otherwise operators still cannot tell what goal/context package the parent actually sends or whether the child silently inherits ambient conversation state.
- Delegated child-agent review also needs an explicit summary schema; otherwise operators still cannot tell what a child report must contain beyond a vague final summary.
- Delegated child-agent review also needs an explicit concurrency cap; otherwise operators still cannot tell how much parallel fan-out one parent task is allowed to create.
- Delegated/background work also needs an explicit notification policy; otherwise operators still cannot tell whether long-running work sends all updates, only results, only errors, or nothing at all.
- Interactive operator control also needs an explicit interrupt contract; otherwise a safe implementation can expose in-flight work without clarifying whether interruption kills commands, cancels queued tools, or merges follow-up prompts.
- Delegated child-agent review also needs an explicit interrupt-propagation rule; otherwise operators still cannot tell whether stopping the parent also stops active children or leaves detached delegated work behind.
- Delegated child-agent review also needs an explicit progress-display rule; otherwise operators still cannot tell whether child work appears as live per-task tool activity, batched parent progress, or some other progress surface.
- Operator-visible tool use also needs an explicit display policy; otherwise operators still cannot tell whether tool activity is hidden, new-only, all, or verbose, or whether that surface can be toggled at runtime.
- Prompt-profile parity also needs an explicit global-identity rule; otherwise operators still cannot tell whether a durable per-user identity file exists outside repo context, where it loads from, or what fallback identity applies when it is missing.
- Repo-context parity also needs an explicit file-type priority rule; otherwise operators still cannot tell whether one project context type wins, how `.hermes.md`, `AGENTS.md`, `CLAUDE.md`, and `.cursorrules` compete, or how that choice stays separate from the global identity layer.
- Long-lived operator workflows also need explicit session-title semantics; otherwise a safe implementation can describe sessions and resumes without saying whether users can name and revisit work by title.
- Resumed-session UX also needs explicit recap semantics; otherwise a safe implementation can describe session resume without saying whether users return to a compact "Previous Conversation" panel or only a one-line resume hint.
- Long-lived titled sessions also need explicit lineage semantics; otherwise a safe implementation can say “resume by name” without telling users whether that means an exact match or the newest session in a lineage.
- Session naming also needs explicit lifecycle semantics; otherwise a safe implementation can mention titles without saying whether they are generated automatically, queued before the first message, or renamed later from a non-chat surface.
- Session persistence also needs explicit exit discoverability semantics; otherwise a safe implementation can support session resume in theory while leaving operators unsure what identifier or command they should use when a session ends.
- Long-lived context also needs explicit session-search semantics; otherwise a safe implementation can promise persistent sessions without saying whether old conversations are searchable or how the agent can recall them later.

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
- If phase 1 adds selection, make sure clear/default semantics appear in the same UX surfaces so operators do not need to edit files just to undo an override.
- If phase 1 adds selection and persistence, decide whether recent bundle changes belong in `status`, `target/` artifacts, or both so operators can debug state transitions later.
- If phase 1 keeps bundle selection separate from memory, state that plainly in status/help/artifacts; if it does not, define the exact interaction before implementation starts.
- If phase 1 keeps separate global-persona and repo-context layers, define the precedence rule early so overlapping instructions are explainable before implementation hardens around the wrong assumption.
- If phase 1 wants Hermes-like repo context, decide early whether the wrapper supports one repo-local file or hierarchical `AGENTS.md` discovery so monorepo behavior does not stay implicit.
- If phase 1 supports hierarchical repo context, decide where merge order and relative path headers show up so operators can debug the assembled prompt instead of only the discovered files.
- If phase 1 wants Hermes-like repo context, decide early whether files like `.cursorrules` are ignored or imported so existing repositories do not lose guidance silently during migration.
- If phase 1 exposes background/delegated tasks, decide early whether they inherit the active profile/model/reasoning settings and where isolation from the parent history is surfaced in status/help/artifacts.
- If phase 1 exposes named bundles, decide early whether they also preload toolsets or skills and where that loaded tooling set becomes visible to the operator.
- If phase 1 exposes background/delegated tasks, decide early whether provider, toolsets, and fallback model inherit too so operators can reason about spawned-task behavior without reading source.
- If phase 1 exposes delegated/background review, decide early whether workers use isolated git worktrees so parallel tasks do not silently share one mutable checkout.
- If phase 1 exposes delegated/background review, decide early where task IDs, progress, and completion/error delivery appear so delegated work stays auditable instead of feeling hidden.
- If phase 1 exposes delegated child-agent review, decide early whether only the final summary re-enters the parent context or whether richer transcripts/artifacts are surfaced separately so parent-context growth stays predictable.
- If phase 1 exposes delegated child-agent review, decide early whether child tools are narrowed to a subset and where the allowed toolsets appear so delegated scope stays explainable.
- If phase 1 exposes delegated child-agent review, decide early whether recursion stops at one child layer and where that limit is surfaced so delegation fan-out stays predictable.
- If phase 1 exposes delegated child-agent review, decide early what goal/context handoff is shown to the operator so child inputs stay auditable instead of implicit.
- If phase 1 exposes delegated child-agent review, decide early what summary fields every child must return so delegated outputs stay comparable instead of free-form.
- If phase 1 exposes delegated child-agent review, decide early what the maximum concurrent child fan-out is and surface that limit so delegated parallelism stays predictable.
- If phase 1 exposes long-running delegated/background work, decide early what notification modes exist and whether completion can ring a bell so background progress stays predictable.
- If phase 1 exposes interactive interruption, decide early whether in-flight commands are killed, queued tools are cancelled, follow-up interruption messages are coalesced, and stop-without-redirect is supported so operator control stays predictable.
- If phase 1 exposes delegated child-agent review, decide early whether parent interruption propagates to active children and surface that cancellation boundary so delegated control flow stays predictable.
- If phase 1 exposes delegated child-agent review, decide early whether progress appears as real-time CLI tree-view activity, batched parent progress updates, or another declared surface so delegated progress stays predictable across interfaces.
- If phase 1 surfaces tool activity, decide early whether visibility is `off`, `new`, `all`, or `verbose`, and whether a runtime toggle like `/verbose` exists so tool-output UX stays predictable across interfaces.
- If phase 1 wants Hermes-like personality behavior, decide early whether a global identity file exists outside repo context, whether it loads only from a home path, and what fallback identity applies so persona layering stays predictable.
- If phase 1 wants Hermes-like context-file behavior, decide early whether project context uses a first-match file-type priority rule and keep that separate from the global identity layer so context selection stays predictable.
- If phase 1 exposes longer-lived sessions, decide early whether users can title, browse, and resume sessions by human-readable names so the operator workflow stays navigable without raw IDs.
- If phase 1 exposes longer-lived sessions, decide early whether resume shows a compact recap panel or only a minimal one-liner, and surface any `resume_display`-style toggle so resumed-session UX stays predictable.
- If phase 1 exposes longer-lived titled sessions, decide early whether resume-by-name targets the newest lineage variant and how compressed/resumed descendants are grouped so named-session workflows stay predictable over time.
- If phase 1 exposes session naming, decide early whether titles are auto-generated, whether `/title` can queue before the first message, and whether rename exists outside chat so title behavior stays predictable across the whole session lifecycle.
- If phase 1 exposes session persistence, decide early whether exit prints the session ID and a direct resume command so the return path to an earlier session stays obvious without extra browsing steps.
- If phase 1 exposes long-lived sessions, decide early whether prior conversations are searchable across sessions and whether a built-in search tool or equivalent full-text recall surface exists so old context can be recovered predictably.
- If phase 1 spans both CLI and messaging, decide early which control commands are shared across interfaces and which remain platform-specific so operators do not have to rediscover the control surface every time they switch entry points.
- If phase 1 exposes reset/new-session flows, decide early what configuration is surfaced at that moment so operators can confirm which model/personality/profile/tooling state the fresh session actually picked up.
- If phase 1 expects operators to reason about active model/provider/profile state mid-session, decide early whether there is an always-visible config/status bar and whether that surface can be toggled without detouring through a separate status command.
- If phase 1 exposes in-flight operator control, decide early whether follow-up prompts may be queued without interruption and when queued input runs, so users do not have to choose blindly between waiting and force-interrupting active work.
