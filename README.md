<p align="center">
  <img src="logo.png" alt="codex-heartbeat logo" width="575">
</p>

# codex-heartbeat

## Problem

Codex still has issues with long-running autorprompting loops in its current setup.

`codex-heartbeat` is a small Go wrapper that is like a Ralph loop but with a few practical knobs. As 
I wanted to help my peers understand and see how to use Codex for long-running autotasking-style loops. Being able to see the session autoprompt but also being able to intervene is valuable.

- https://x.com/karpathy/status/2031083551387701698

It does four things:

- Starts a session once and persists the explicit session id.
- Resumes that same session on a timer instead of relying on `resume --last`.
- Holds a workspace lock so two heartbeat runners do not manage the same thread at once.
- Tees Codex output to the terminal while also appending it to daily log files under `~/.codex-heartbeat/projects/<workspace-key>/logs/`.

## Warning 

> Warning
> codex-heartbeat runs Codex in the current `--yolo`-equivalent mode by default.
> On this Codex CLI version that means it automatically passes `--dangerously-bypass-approvals-and-sandbox` to child Codex runs unless you add `--safe`.


## Build

```bash
go build ./cmd/codex-heartbeat
```

## Basic usage

Initialize an autoresearch-style workspace once:

```bash
./codex-heartbeat init --workdir /path/to/workdir
```

That scaffolds:

- `program.md`
- `PLANNING.md`
- `target/results.jsonl`
- `target/latest-context.md`
- `target/templates/{plan,execution,results,insights}.md`

The normal loop is:

1. The human edits `program.md`.
2. The agent works in the target workspace.
3. `codex-heartbeat` keeps re-injecting a bounded experiment-loop prompt and points the agent at the run artifacts under `target/`.
4. The agent writes memory into `target/run-<timestamp>/` and `target/results.jsonl`.

Minimal `program.md` example:

```md
# Program

Objective: Fix the current failure with the smallest safe change.
Primary evaluator: go test ./...
Prompt mode: autoresearch
Council after failures: 3
Checkpoint commits: true

## Notes

- Keep one hypothesis per cycle.
- Record keep, discard, or revert in target/results.jsonl.
```

Prompt precedence is:

1. `--prompt`
2. repo-local `program.md`
3. the embedded fallback prompt template

`--prompt` is a full override. When it is present, codex-heartbeat re-reads that file on every emission, refreshes the cached copy on success, and falls back to the cached copy if the file later disappears. When `--prompt` is absent, codex-heartbeat re-reads `program.md` on every emission and renders the embedded autoresearch loop template around it. If neither exists, it renders the embedded fallback template by itself.

Bootstrap or pulse once:

```bash
./codex-heartbeat pulse --workdir /path/to/workdir
```

Open the interactive Codex UI with transcript logging and the default screen-aware heartbeat mode:

```bash
./codex-heartbeat run --workdir /path/to/workdir
```

Keep the wrapper banner and Codex output in normal scrollback:

```bash
./codex-heartbeat run --workdir /path/to/workdir --no-alt-screen
```

Use an explicit prompt override instead of `program.md`:

```bash
./codex-heartbeat run --workdir /path/to/workdir --prompt /path/to/workdir/heartbeat.md --interval 15m
```

Explicitly select the default screen-aware mode, which looks for 15 seconds of idle screen state, waits for 20 seconds of quiet local input, and falls back to a heartbeat after 60 minutes without an injected prompt:

```bash
./codex-heartbeat run --workdir /path/to/workdir --screen-idle-heartbeat
```

Run the interactive heartbeat for a bounded amount of time:

```bash
./codex-heartbeat run --workdir /path/to/workdir --interval 15m --end-in 2 hours
```

Run the unattended heartbeat loop:

```bash
./codex-heartbeat daemon --workdir /path/to/workdir --interval 5m
```

Stop the unattended heartbeat after one day:

```bash
./codex-heartbeat daemon --workdir /path/to/workdir --interval 5m --end-in 1 day
```

Inspect the stored session:

```bash
./codex-heartbeat status --workdir /path/to/workdir
```

Example autoresearch programs ship in:

- `examples/program-debugging.md`
- `examples/program-benchmark.md`
- `examples/program-manual-validation.md`

## Autoresearch Model

The default embedded prompt is no longer a generic “keep going” reminder. It is an experiment/debug loop contract that tells Codex to:

- work on a single objective
- reuse a single primary evaluator
- choose one hypothesis per cycle
- explicitly keep, discard, or revert
- read memory from `target/latest-context.md`
- write memory into `target/run-<timestamp>/` and `target/results.jsonl`

The 3-agent council is now a fallback, not the default first move. The prompt tells Codex to use the council only when it is blocked or when the recent failure streak in `target/results.jsonl` reaches the configured threshold from `program.md`.

`Prompt mode: manual-test-first` is a guidance mode for workflows where Codex should prepare the next candidate fix and validation steps, then stop before the final human gate.

## Recommended Workspace Contract

For long-running loops, a target workspace should ideally contain:

- `program.md`: human-authored objective, evaluator, and constraints
- `PLANNING.md`: optional broader notes or backlog
- `AGENTS.md`: repo-local rules for the agent
- `target/run-<timestamp>/insights.md`: concise memory from each run
- `target/results.jsonl`: compact experiment ledger
- `target/latest-context.md`: bounded summary used on the next heartbeat

## State and logs

By default the wrapper writes:

- `~/.codex-heartbeat/projects/<workspace-key>/state.json`
- `~/.codex-heartbeat/projects/<workspace-key>/screen-state.json`
- `~/.codex-heartbeat/projects/<workspace-key>/heartbeat.lock`
- `~/.codex-heartbeat/projects/<workspace-key>/prompts/<hash>.txt`
- `~/.codex-heartbeat/projects/<workspace-key>/logs/YYYY-MM-DD.jsonl`
- `~/.codex-heartbeat/projects/<workspace-key>/logs/YYYY-MM-DD-screen.jsonl`
- `~/.codex-heartbeat/projects/<workspace-key>/logs/YYYY-MM-DD-run.log`

Autoresearch mode also writes bounded memory into the target workspace:

- `<workdir>/target/results.jsonl`
- `<workdir>/target/latest-context.md`
- `<workdir>/target/run-<timestamp>/{plan,execution,results,insights}.md`

`<workspace-key>` is derived from the selected `--workdir` so each workspace gets its own runtime state under your home directory. If an older `<workdir>/.codex-heartbeat` directory exists and you are using the default paths, the wrapper moves it into `~/.codex-heartbeat/projects/` automatically.

The session id is discovered after bootstrap by scanning `$CODEX_HOME/sessions` or `~/.codex/sessions` for the newest `session_meta` record whose `cwd` matches the selected workspace.

## Notes

- `run` acquires the workspace lock, attaches you to the live Codex UI, and keeps a transcript log in the background.
- `run` now defaults to the screen-aware scheduler. It polls the live Codex screen every 5 seconds, treats 3 consecutive idle polls as idle readiness, waits for 20 seconds of quiet local input before injecting, and falls back to a heartbeat after 60 minutes without an injected prompt.
- Fresh interactive boots leave Codex alone for about 5 seconds before sending the initial prompt so the UI can settle first.
- `run --interval 15m` keeps the live UI attached but switches to an explicit timed scheduler. On resume it injects one heartbeat after a short startup settle delay, then continues on the configured interval.
- `--screen-idle-heartbeat` is still accepted as an explicit alias for the default screen-aware scheduler.
- `--interval` and `--screen-idle-heartbeat` are mutually exclusive because they choose different heartbeat schedulers.
- `run` prints a short `codex-heartbeat` banner before attach. In Ghostty on macOS it defaults to inline mode so that banner stays visible; use `--alt-screen` to force the alternate screen or `--no-alt-screen` on other terminals when you want the same inline behavior.
- `run` also sets the terminal title to `codex-heartbeat | <workdir>` so heartbeat tabs are easy to spot at a glance.
- `--interval` and `--end-in` accept short and long units for minutes, hours, and days such as `30m`, `2h`, `1d`, `15 minutes`, `2 hours`, and `1 day`.
- The screen-aware scheduler watches Codex's status line for active indicators such as `Working (3m 02s • esc to interrupt)`, deliberately waits on ambiguous screens, keeps tracking idle screen state even while the recent-input guard is active, and only injects once local input has been quiet for 20 seconds.
- The latest screen classifier snapshot is written to `screen-state.json`, and every screen poll is appended to `YYYY-MM-DD-screen.jsonl` so you can audit why the wrapper thought Codex was working, idle, ambiguous, or blocked by recent input.
- Prompt emissions now resolve prompt sources in this order: `--prompt`, then repo-local `program.md`, then the embedded fallback template.
- Explicit `--prompt` files are re-read on every send. Successful reads refresh the workspace cache, and missing prompt files fall back to that cached copy; if neither exists, the run fails.
- `program.md` is also re-read on every send so the human can change the research/debugging program while the agent works in the target workspace.
- The default prompt writes bounded memory into `target/` and treats the 3-agent council as a fallback after repeated failed cycles, not as the default first action on every heartbeat.
- If no tracked session id exists yet, `run` starts a brand-new interactive Codex session using the prompt file and then persists the discovered session id afterward.
- `daemon` is the old timed heartbeat loop for unattended runs.
- `pulse` and `bootstrap` acquire the same lock for a single execution and skip if another process already holds it.
- `pulse`, `bootstrap`, and `daemon` enable `--skip-git-repo-check` by default so non-interactive child runs can target a plain directory.
- Child exit codes are preserved for wrapped one-shot and interactive runs such as `pulse`, `bootstrap`, and `run`.
- `run` now uses a PTY and terminal raw mode so you can keep using the Codex UI while codex-heartbeat logs the transcript and injects prompts when the scheduler decides it is time.
