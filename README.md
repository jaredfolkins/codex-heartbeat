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

Put a standing prompt in your target workspace, for example `heartbeat.md`:

```text
IF you are already in the middle of an active step, finish that step before changing direction.
Continue the current task from where you left off.
Work in bounded steps.
Update notes/files in this workspace.
Where appropriate, use sub-agents.
If blocked, write the blocker clearly and propose the next action.
Do not restart from scratch.
```

`--prompt` is optional. If you do not provide it, codex-heartbeat uses the embedded `heartbeat.md` that ships inside the binary.

Bootstrap or pulse once:

```bash
./codex-heartbeat pulse --workdir /path/to/workdir
```

Open the interactive Codex UI with transcript logging:

```bash
./codex-heartbeat run --workdir /path/to/workdir
```

Keep the wrapper banner and Codex output in normal scrollback:

```bash
./codex-heartbeat run --workdir /path/to/workdir --no-alt-screen
```

Open the interactive Codex UI and auto-paste the prompt file every 15 minutes:

```bash
./codex-heartbeat run --workdir /path/to/workdir --prompt /path/to/workdir/heartbeat.md --interval 15m
```

Run the interactive heartbeat for a bounded amount of time:

```bash
./codex-heartbeat run --workdir /path/to/workdir --interval 15m --end-in 2 hours
```

Run the old unattended heartbeat loop:

```bash
./codex-heartbeat daemon --workdir /path/to/workdir --prompt /path/to/workdir/heartbeat.md --interval 5m
```

Stop the unattended heartbeat after one day:

```bash
./codex-heartbeat daemon --workdir /path/to/workdir --prompt /path/to/workdir/heartbeat.md --interval 5m --end-in 1 day
```

Inspect the stored session:

```bash
./codex-heartbeat status --workdir /path/to/workdir
```

## State and logs

By default the wrapper writes:

- `~/.codex-heartbeat/projects/<workspace-key>/state.json`
- `~/.codex-heartbeat/projects/<workspace-key>/heartbeat.lock`
- `~/.codex-heartbeat/projects/<workspace-key>/logs/YYYY-MM-DD.jsonl`
- `~/.codex-heartbeat/projects/<workspace-key>/logs/YYYY-MM-DD-run.log`

`<workspace-key>` is derived from the selected `--workdir` so each workspace gets its own runtime state under your home directory. If an older `<workdir>/.codex-heartbeat` directory exists and you are using the default paths, the wrapper moves it into `~/.codex-heartbeat/projects/` automatically.

The session id is discovered after bootstrap by scanning `$CODEX_HOME/sessions` or `~/.codex/sessions` for the newest `session_meta` record whose `cwd` matches the selected workspace.

## Notes

- `run` acquires the workspace lock, attaches you to the live Codex UI, and keeps a transcript log in the background.
- `run --interval 15m` keeps the live UI attached. On resume it injects one heartbeat after a short startup settle delay, then continues on the configured interval.
- `run` prints a short `codex-heartbeat` banner before attach. In Ghostty on macOS it defaults to inline mode so that banner stays visible; use `--alt-screen` to force the alternate screen or `--no-alt-screen` on other terminals when you want the same inline behavior.
- `--interval` and `--end-in` accept short and long units for minutes, hours, and days such as `30m`, `2h`, `1d`, `15 minutes`, `2 hours`, and `1 day`.
- If no tracked session id exists yet, `run` starts a brand-new interactive Codex session using the prompt file and then persists the discovered session id afterward.
- `daemon` is the old timed heartbeat loop for unattended runs.
- `pulse` and `bootstrap` acquire the same lock for a single execution and skip if another process already holds it.
- `pulse`, `bootstrap`, and `daemon` enable `--skip-git-repo-check` by default so non-interactive child runs can target a plain directory.
- Child exit codes are preserved for wrapped one-shot and interactive runs such as `pulse`, `bootstrap`, and `run`.
- `run` now uses a PTY and terminal raw mode so you can keep using the Codex UI while codex-heartbeat logs the transcript and injects timed prompts.
