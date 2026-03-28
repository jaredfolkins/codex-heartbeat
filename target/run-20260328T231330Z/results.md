# Results

- Status: success
- Council triggered at start: false
- Primary evaluator: `rg -n "^### Hermes Parity Gap|stronger launch-time instruction channel|ephemeral prefill|harmless canary-scoring harness|parity claim rule" PLANNING.md`

## Observable Signals

- `PLANNING.md` now contains a concrete `[ ]` checklist for a safe prompt-profile feature derived from the Hermes Godmode architecture.
- The evaluator confirmed the key upstream gap: `codex-heartbeat` currently launches `codex` / `codex resume` through `buildInteractiveArgs()`, while upstream Codex SDK and app-server surfaces expose `base_instructions`, `developer_instructions`, and `config.model_reasoning_effort`.
- Hermes's "Godmode" implementation depends on system prompt injection, ephemeral prefill, model-family strategy selection, and canary scoring. Those are stronger primitives than the current heartbeat wrapper's user-message reinjection path.
- The exact X post text was not dependable enough to use as the primary source, but the Hermes repo supplied enough concrete implementation detail to build the checklist.
- `PLANNING.md` now also separates implementation tasks, blocked/non-goals, and acceptance criteria, which makes the next implementation pass less ambiguous.
- The planning-structure evaluator found all three sections plus the checkbox items in one place.
- `PLANNING.md` now also includes a `Phase 1 Recommendation`, so the next safe implementation pass has a clear starting slice instead of only a broad backlog.
- `codex-heartbeat run` now accepts wrapper-safe `--profile`, `--model`, and `--model-reasoning-effort` flags and forwards them to the child Codex CLI in upstream-compatible syntax.
- `buildInteractiveArgs()` now emits `--profile`, `--model`, and `--config model_reasoning_effort="high"` before launching or resuming Codex.
- The focused evaluator passed, covering both direct arg-building and a fake-child launch path.
- The function still does not appear to be the same as Hermes Agent: the wrapper now exposes phase-1 launch selection, but it still lacks Hermes-style base/developer instruction injection, ephemeral prefill, and canary scoring.
- `PLANNING.md` now contains a dedicated `Hermes Parity Gap` checklist, so the repo records the exact remaining conditions that keep the parity answer at "no".

## Disposition

- keep
