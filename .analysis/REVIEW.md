# Review Notes

Date: 2026-03-21

Recovered task: continue the interrupted review of the current `codex-heartbeat` changes.

Status: review completed and follow-up fixes applied in the working tree.

## Fixed In Source

1. High: restored the original new-session behavior for `run --interval`. A brand-new interactive session now starts with the prompt file again, while immediate heartbeat injection remains resume-only.

2. Medium: corrected the CLI/help surface for `--skip-git-repo-check`. Local verification showed the upstream Codex CLI only accepts that flag for `codex exec`, not interactive `codex`, so `run` and `status` no longer expose it.

3. Medium: normalized subcommand help handling so `pulse/run/daemon/status --help` exit successfully instead of returning `flag: help requested`.

4. Low: cleaned up README drift while touching the affected paths, including stale `codex-loop` references, the `problem i'm solving` typo, and the non-interactive-only explanation for `--skip-git-repo-check`.

5. Medium: restored automatic migration of an old `<workdir>/.codex-heartbeat` runtime directory into `~/.codex-heartbeat/projects/`, including a copy/delete fallback for cross-device moves, and made `status` use the same migration path.

## Remaining Notes

1. No additional review findings remain from this pass.

## Workspace Status

1. Ready for commit review. Current tracked modifications are `README.md`, `cmd/codex-heartbeat/main.go`, and `cmd/codex-heartbeat/main_test.go`.

2. This review note lives at `.analysis/REVIEW.md` and is currently untracked.

## Handoff

1. Suggested staged files:
   `README.md`, `cmd/codex-heartbeat/main.go`, `cmd/codex-heartbeat/main_test.go`, `.analysis/REVIEW.md`

2. Suggested commit title:
   `Fix heartbeat run bootstrap and CLI runtime UX`

3. Suggested commit summary:
   restore new-session `run --interval` prompt startup, normalize subcommand help exits, limit `--skip-git-repo-check` to non-interactive commands, add legacy runtime migration coverage, and refresh README/docs.

4. Suggested next commands:
   `git add README.md cmd/codex-heartbeat/main.go cmd/codex-heartbeat/main_test.go .analysis/REVIEW.md`
   `git commit -m "Fix heartbeat run bootstrap and CLI runtime UX"`

## Validation

- `go test ./...`
- `go test -race ./...`
- `go vet ./...`
- targeted fake-`codex` test covering new-session `run --interval` prompt launch behavior
- `go run ./cmd/codex-heartbeat run --help`
- `go run ./cmd/codex-heartbeat status --help`
- `./codex-heartbeat run --help`
- `./codex-heartbeat status --help`
- `./codex-heartbeat pulse --help`
- `./codex-heartbeat daemon --help`
- temp-workspace integration check confirming `status` migrates legacy `.codex-heartbeat/state.json`
- post-fix grep sweep confirming stale naming/help text references only remain in intentional tests and review notes
