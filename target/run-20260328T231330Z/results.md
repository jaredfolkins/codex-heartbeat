# Results

- Status: success
- Council triggered at start: false
- Primary evaluator: `go test ./cmd/codex-heartbeat -run 'StatusCommandIncludesProgramLaunchSettings|PromptResolverWritesLaunchSettingsToLatestContext|RecordRunStartWritesEvaluatorToResultsLedger|LoadProgramConfigParsesLaunchOverrides|RegisterRunFlags|RunInteractiveCommandPassesLaunchOverrides' -count=1`

## Observable Signals

- `codex-heartbeat status` now includes a `launch_settings` object when `program.md` resolves a profile/model/reasoning-effort.
- The existing artifact evidence remains in place: `latest-context.md` and the pending run-start ledger note still record the same launch summary.
- The actual child Codex launch behavior stayed stable: the wrapper still emits `--profile`, `--model`, and `--config model_reasoning_effort="high"` once those values are present in `program.md`.
- The focused evaluator passed for `status` JSON, latest-context evidence, run-start ledger evidence, metadata parsing, absent wrapper flags, and fake-child launch arg forwarding.
- The function still does not appear to be the same as Hermes Agent because the wrapper still lacks stronger launch-time instruction control, ephemeral prefill, and benign canary scoring.

## Disposition

- keep
