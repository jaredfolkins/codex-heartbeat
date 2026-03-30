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

The first `run` against a workspace now auto-scaffolds the autoresearch files when they are missing. If the scaffold is completely absent and the terminal is interactive, `codex-heartbeat` asks a short init questionnaire for the goal, evaluator, first deep dive, and starting prompt mode before writing the files. This is currently a conversational questionnaire, not a full-screen TUI:

```bash
./codex-heartbeat --workdir /path/to/workdir
```

That creates, when absent:

- `program.md`
- `PLANNING.md`
- `target/PLANNING_HISTORY.md`
- `target/results.jsonl`
- `target/latest-context.md`
- `target/templates/{plan,execution,results,insights}.md`

The normal loop is:

1. The human edits `program.md`.
2. The agent keeps active tasks in `PLANNING.md` and archives completed or superseded plan items into `target/PLANNING_HISTORY.md`.
3. The agent works in the target workspace.
4. `codex-heartbeat` keeps re-injecting a bounded experiment-loop prompt and points the agent at the run artifacts under `target/`.
5. The agent writes memory into `target/run-<timestamp>/` and `target/results.jsonl`.
6. When the objective is fully achieved, the agent creates `agent-paused.lock` so future heartbeat injections stop until the file is removed.

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

1. repo-local `program.md`
2. the embedded fallback prompt template

`codex-heartbeat` now follows the repo-local workflow directly. It re-reads `program.md` on every emission and renders the embedded autoresearch loop template around it. If `program.md` is missing or unusable, it falls back to the embedded template defaults.

Open the interactive Codex UI with transcript logging and the default screen-aware heartbeat mode:

```bash
./codex-heartbeat --workdir /path/to/workdir
```

`./codex-heartbeat run --workdir /path/to/workdir` is equivalent. `run` remains available, but bare flags default to `run`.

Keep the wrapper banner and Codex output in normal scrollback:

```bash
./codex-heartbeat --workdir /path/to/workdir --no-alt-screen
```

Use the council repeatedly during autoresearch instead of only as a stuck-state fallback:

```bash
./codex-heartbeat --workdir /path/to/workdir --council
```

Set a safe launch profile, model, and reasoning-effort selection in `program.md`:

```bash
cat >> /path/to/workdir/program.md <<'EOF'
Profile: safe-research
Model: gpt-5.3-codex-spark
Model reasoning effort: high
EOF
```

This is a phase-1 prompt-profile feature, not Hermes parity. Today it only forwards wrapper-safe launch settings to the Codex CLI. It does not yet provide Hermes-style base/developer instruction injection, ephemeral prefill, or canary scoring.

Explicitly select the default screen-aware mode, which looks for 15 seconds of idle screen state, waits for 20 seconds of quiet local input, and falls back to a heartbeat after 60 minutes without an injected prompt:

```bash
./codex-heartbeat --workdir /path/to/workdir --screen-idle-heartbeat
```

Run the interactive heartbeat for a bounded amount of time:

```bash
./codex-heartbeat --workdir /path/to/workdir --interval 15m --end-in 2 hours
```

Inspect the stored session:

```bash
./codex-heartbeat status --workdir /path/to/workdir
```

`status` now reports the current session state plus two autoresearch-specific summaries when available:

- `launch_settings`: the resolved `program.md` profile, model, and reasoning effort.
- `hermes_parity`: whether the current wrapper surface is equivalent to Hermes Agent and, when it is not, the concrete missing capabilities plus a safe `task_list` of next steps, a `claim_rule` for when parity must remain false, and a `review_basis` list pointing at the reviewed Hermes sources.

Example autoresearch programs ship in:

- `examples/program-planning.md`
- `examples/program-debugging.md`
- `examples/program-benchmark.md`
- `examples/program-manual-validation.md`

The source Markdown templates used by the wrapper live in `cmd/codex-heartbeat/templates/`.

`codex-heartbeat --help` prints the detailed guide with examples and ideas. Add `--brevity` after `--help` when you want the compact help instead, for example `codex-heartbeat --help --brevity` or `codex-heartbeat run --help --brevity`.

## Autoresearch Model

The default embedded prompt is no longer a generic “keep going” reminder. It is an experiment/debug loop contract that tells Codex to:

- work on a single objective
- reuse a single primary evaluator
- choose one hypothesis per cycle
- explicitly keep, discard, or revert
- read memory from `target/latest-context.md`
- keep active work in `PLANNING.md`
- preserve completed or superseded planning items in `target/PLANNING_HISTORY.md`
- pause the loop with `agent-paused.lock` once the objective is fully achieved
- write memory into `target/run-<timestamp>/` and `target/results.jsonl`

By default the 3-agent council is still a fallback, not the default first move. The prompt tells Codex to use the council only when it is blocked or when the recent failure streak in `target/results.jsonl` reaches the configured threshold from `program.md`.

When you pass `--council`, the prompt switches to a frequent-council mode: use the council during baseline framing, next-hypothesis selection, and post-evaluator interpretation. The guidance is to keep the root agent on `gpt-5.4` with `xhigh` reasoning and use `gpt-5.3-codex-spark` with `high` reasoning for the three sub-agents.

`Prompt mode: planning` is a guidance mode for using the autoresearch loop to refine the goal, deepen `PLANNING.md`, and decide the next deep dive before broad implementation. On the first planning-mode run in a workspace, the prompt now starts with an adversarial 3-agent planning council for several rounds before it settles the initial plan.

`Prompt mode: manual-test-first` is a guidance mode for workflows where Codex should prepare the next candidate fix and validation steps, then stop before the final human gate.

If `agent-paused.lock` exists in the workspace root, heartbeat injections do not fire. The wrapper also auto-creates that file when the latest `target/results.jsonl` entry records a completion disposition such as `complete`. Remove the file when you want the loop to resume.

## Launch Profiles

`codex-heartbeat` now exposes a small wrapper-safe launch-profile layer through `program.md` metadata:

- `Profile: NAME` forwards Codex's config profile selection.
- `Model: NAME` forwards an explicit model choice.
- `Model reasoning effort: LEVEL` forwards `--config model_reasoning_effort="LEVEL"` to the child Codex CLI.

This is enough to make model/profile selection reproducible in `codex-heartbeat`, but it is still not equivalent to Hermes Agent's `godmode` design. The wrapper does not yet control stronger launch-time instruction channels such as `base_instructions`, `developer_instructions`, ephemeral prefill, a benign canary-scoring harness, or a Hermes-style delegated cross-review workflow for research loops.

## Recommended Workspace Contract

For long-running loops, a target workspace should ideally contain:

- `program.md`: human-authored objective, evaluator, and constraints
- `PLANNING.md`: the live active plan
- `agent-paused.lock`: optional pause sentinel that disables future heartbeat injections until removed
- `target/PLANNING_HISTORY.md`: durable planning memory for completed or superseded work
- `AGENTS.md`: repo-local rules telling the agent which files to read and what each one owns
- `target/run-<timestamp>/insights.md`: concise memory from each run
- `target/results.jsonl`: compact experiment ledger
- `target/latest-context.md`: bounded summary used on the next heartbeat

## State and logs

By default the wrapper writes:

- `~/.codex-heartbeat/projects/<workspace-key>/state.json`
- `~/.codex-heartbeat/projects/<workspace-key>/screen-state.json`
- `~/.codex-heartbeat/projects/<workspace-key>/heartbeat.lock`
- `~/.codex-heartbeat/projects/<workspace-key>/logs/YYYY-MM-DD.jsonl`
- `~/.codex-heartbeat/projects/<workspace-key>/logs/YYYY-MM-DD-screen.jsonl`
- `~/.codex-heartbeat/projects/<workspace-key>/logs/YYYY-MM-DD-run.log`

Autoresearch mode also writes bounded memory into the target workspace:

- `<workdir>/target/PLANNING_HISTORY.md`
- `<workdir>/target/results.jsonl`
- `<workdir>/target/latest-context.md`
- `<workdir>/target/run-<timestamp>/{plan,execution,results,insights}.md`

`<workspace-key>` is derived from the selected `--workdir` so each workspace gets its own runtime state under your home directory. If an older `<workdir>/.codex-heartbeat` directory exists and you are using the default paths, the wrapper moves it into `~/.codex-heartbeat/projects/` automatically.

The session id is discovered after startup by scanning `$CODEX_HOME/sessions` or `~/.codex/sessions` for the newest `session_meta` record whose `cwd` matches the selected workspace.

## Heartbeat Detection: Triage And Validation

### For operators

- Think of the screen-aware scheduler as a guarded workflow: it prefers a live active status row over stale transcript text, waits on ambiguous screens, and only injects after local input has been quiet for 20 seconds.
- If heartbeat did not fire when you expected, inspect `~/.codex-heartbeat/projects/<workspace-key>/screen-state.json` first. Check `screen_state`, `reason`, `quiet`, `idle_polls`, and `should_inject` to see whether Codex still looked active, recent local input was blocking injection, or the screen stayed ambiguous.
- Tail `~/.codex-heartbeat/projects/<workspace-key>/logs/YYYY-MM-DD-screen.jsonl` when you need a timeline instead of just the latest state. That log shows each poll decision and is the fastest way to confirm whether the wrapper saw `working`, `idle`, or `ambiguous` before it injected.
- If heartbeat fired unexpectedly, compare the recent screen log against known false-positive traps such as queued-message previews, footer-only background-terminal text, or historical background-terminal output. The fixture corpus in `cmd/codex-heartbeat/testdata/screen/` shows the kinds of snapshots the detector is expected to classify safely.
- If the council triggered unexpectedly in the default mode, inspect the trailing dispositions in `<workdir>/target/results.jsonl`. Council fallback follows consecutive failure-like dispositions, not a single bad outcome.
- If you passed `--council`, expect the council to appear throughout the autoresearch loop. In that mode the question is whether the loop stayed bounded around one hypothesis and one evaluator, not whether the council appeared at all.
- When a run feels off, inspect the latest `<workdir>/target/run-<timestamp>/{plan,execution,results,insights}.md` alongside `target/latest-context.md`. Those files show what the loop believed it was doing and whether the recent failure streak or evaluator history explains the current behavior.

### For contributors

- Reproduce detector behavior without attaching to a live PTY first. The focused verification path is `go test ./cmd/codex-heartbeat -run 'Screen|Replay' -count=1`.
- Reproduce autoresearch and council-threshold behavior with `go test ./cmd/codex-heartbeat -run 'Autoresearch|Council|ResultLedger|ShouldTriggerCouncil' -count=1`.
- The fixture-backed screen evidence lives in `cmd/codex-heartbeat/testdata/screen/`, and the replay-oriented detector coverage lives in `cmd/codex-heartbeat/screen_replay_test.go`.
- When docs and behavior diverge, update the README, the relevant fixtures, and the focused tests together. The docs should describe observable outcomes and evidence paths, not freeze every detector heuristic as a permanent contract.

## Notes

- `run` acquires the workspace lock, attaches you to the live Codex UI, and keeps a transcript log in the background. Bare flags such as `codex-heartbeat --workdir /repo` resolve to `run`.
- `run` now defaults to the screen-aware scheduler. It polls the live Codex screen every 5 seconds, treats 3 consecutive idle polls as idle readiness, waits for 20 seconds of quiet local input before injecting, and falls back to a heartbeat after 60 minutes without an injected prompt.
- Before `run` resolves prompts, it checks for the autoresearch scaffold in the selected workspace. If the scaffold is entirely missing, it creates it automatically. If the scaffold is partial, it warns and creates only the missing files without overwriting the existing ones.
- Fresh interactive boots leave Codex alone for about 5 seconds before sending the initial prompt so the UI can settle first.
- `run --interval 15m` keeps the live UI attached but switches to an explicit timed scheduler. On resume it injects one heartbeat after a short startup settle delay, then continues on the configured interval.
- `--screen-idle-heartbeat` is still accepted as an explicit alias for the default screen-aware scheduler.
- `--interval` and `--screen-idle-heartbeat` are mutually exclusive because they choose different heartbeat schedulers.
- `--council` changes the autoresearch prompt policy from fallback-council to frequent-council mode. It does not remove the one-hypothesis/one-evaluator discipline.
- `run` prints a short `codex-heartbeat` banner before attach. In Ghostty on macOS it defaults to inline mode so that banner stays visible; use `--alt-screen` to force the alternate screen or `--no-alt-screen` on other terminals when you want the same inline behavior.
- `run` also sets the terminal title to `codex-heartbeat | <workdir>` so heartbeat tabs are easy to spot at a glance.
- `--interval` and `--end-in` accept short and long units for minutes, hours, and days such as `30m`, `2h`, `1d`, `15 minutes`, `2 hours`, and `1 day`.
- The screen-aware scheduler watches Codex's status line for active indicators such as `Working (3m 02s • esc to interrupt)`, deliberately waits on ambiguous screens, keeps tracking idle screen state even while the recent-input guard is active, and only injects once local input has been quiet for 20 seconds. See `Heartbeat Detection: Triage And Validation` above for the artifact-driven troubleshooting flow.
- The latest screen classifier snapshot is written to `screen-state.json`, and every screen poll is appended to `YYYY-MM-DD-screen.jsonl` so you can audit why the wrapper thought Codex was working, idle, ambiguous, or blocked by recent input.
- Prompt emissions now resolve prompt sources in this order: repo-local `program.md`, then the embedded fallback template.
- `program.md` is also re-read on every send so the human can change the research/debugging program while the agent works in the target workspace.
- The default prompt writes bounded memory into `target/` and treats the 3-agent council as a fallback after repeated failed cycles, not as the default first action on every heartbeat. `--council` opts into a more collaborative version of the loop where the council is used at multiple decision points.
- If no tracked session id exists yet, `run` starts a brand-new interactive Codex session using the `program.md`-driven prompt and then persists the discovered session id afterward.
- Child exit codes are preserved for the wrapped interactive `run`.
- `run` now uses a PTY and terminal raw mode so you can keep using the Codex UI while codex-heartbeat logs the transcript and injects prompts when the scheduler decides it is time.
