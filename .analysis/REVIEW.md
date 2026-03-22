# Review Notes

Date: 2026-03-21

Recovered task: continue the interrupted review of the current `codex-heartbeat` changes.

Status: review completed, follow-up fixes committed, and the branch is now one commit ahead of `origin/main`.

## Fixed In Source

1. High: restored the original new-session behavior for `run --interval`. A brand-new interactive session now starts with the prompt file again, while immediate heartbeat injection remains resume-only.

2. Medium: corrected the CLI/help surface for `--skip-git-repo-check`. Local verification showed the upstream Codex CLI only accepts that flag for `codex exec`, not interactive `codex`, so `run` and `status` no longer expose it.

3. Medium: normalized subcommand help handling so `pulse/run/daemon/status --help` exit successfully instead of returning `flag: help requested`.

4. Low: cleaned up README drift while touching the affected paths, including stale `codex-loop` references, the `problem i'm solving` typo, and the non-interactive-only explanation for `--skip-git-repo-check`.

5. Medium: restored automatic migration of an old `<workdir>/.codex-heartbeat` runtime directory into `~/.codex-heartbeat/projects/`, including a copy/delete fallback for cross-device moves, and made `status` use the same migration path.

## Remaining Notes

1. No additional review findings remain from this pass.

## Workspace Status

1. Committed locally as `eb7ed01` with title `Fix heartbeat run bootstrap and CLI runtime UX`.

2. The working tree is currently clean. `git status --short` returned no modified or untracked files.

3. The local branch is `main` and is currently `ahead 1` relative to `origin/main`.

4. This repository does not currently use `TASKS.md`, `TASKS-HISTORY.md`, `DESIGN.md`, or `AGENTS.md`; the review/handoff note in `.analysis/REVIEW.md` is the active workspace ledger for this task.

## Handoff

1. Latest local commit:
   `eb7ed01 Fix heartbeat run bootstrap and CLI runtime UX`

2. Commit summary:
   restore new-session `run --interval` prompt startup, normalize subcommand help exits, limit `--skip-git-repo-check` to non-interactive commands, add legacy runtime migration coverage, and refresh README/docs.

3. Suggested next command:
   `git push origin main`

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
