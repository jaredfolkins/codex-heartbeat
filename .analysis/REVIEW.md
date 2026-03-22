# Review Notes

Date: 2026-03-21
Last verified: 2026-03-22

Recovered task: continue the interrupted review of the current `codex-heartbeat` changes.

Status: review completed, the follow-up fixes were validated and pushed to `origin/main`, and later synced commits only adjusted documentation wording.

## Fixed In Source

1. High: restored the original new-session behavior for `run --interval`. A brand-new interactive session now starts with the prompt file again, while immediate heartbeat injection remains resume-only.

2. Medium: corrected the CLI/help surface for `--skip-git-repo-check`. Local verification showed the upstream Codex CLI only accepts that flag for `codex exec`, not interactive `codex`, so `run` and `status` no longer expose it.

3. Medium: normalized subcommand help handling so `pulse/run/daemon/status --help` exit successfully instead of returning `flag: help requested`.

4. Low: cleaned up README drift while touching the affected paths, including stale `codex-loop` references, the `problem i'm solving` typo, and the non-interactive-only explanation for `--skip-git-repo-check`.

5. Medium: restored automatic migration of an old `<workdir>/.codex-heartbeat` runtime directory into `~/.codex-heartbeat/projects/`, including a copy/delete fallback for cross-device moves, and made `status` use the same migration path.

## Remaining Notes

1. No additional review findings remain from this pass.

## Workspace Status

1. Primary code change landed in commit `eb7ed01` with title `Fix heartbeat run bootstrap and CLI runtime UX`.

2. This repository does not currently use `TASKS.md`, `TASKS-HISTORY.md`, `DESIGN.md`, or `AGENTS.md`; the review/handoff note in `.analysis/REVIEW.md` is the active workspace ledger for this task.

3. Commits after `eb7ed01` in the current branch history are review-note bookkeeping or README-only documentation edits; they do not reopen the runtime behavior addressed by the fix set.

4. At the latest verification pass on 2026-03-22, local `main` and `origin/main` were synced and the worktree was clean.

## Handoff

1. Primary implementation commit:
   `eb7ed01 Fix heartbeat run bootstrap and CLI runtime UX`

2. Commit summary:
   restore new-session `run --interval` prompt startup, normalize subcommand help exits, limit `--skip-git-repo-check` to non-interactive commands, add legacy runtime migration coverage, and refresh README/docs.

3. Later commits in this branch history after the primary implementation are review-note or README documentation changes only.

4. No additional action is required for this task. If work resumes later, start by checking `git status --short --branch`.

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
