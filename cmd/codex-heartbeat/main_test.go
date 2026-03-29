package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestParseFlexibleDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  time.Duration
	}{
		{input: "30m", want: 30 * time.Minute},
		{input: "2h", want: 2 * time.Hour},
		{input: "1d", want: 24 * time.Hour},
		{input: "15 minute", want: 15 * time.Minute},
		{input: "15 minutes", want: 15 * time.Minute},
		{input: "2 hour", want: 2 * time.Hour},
		{input: "2 hours", want: 2 * time.Hour},
		{input: "1 day", want: 24 * time.Hour},
		{input: "3 days", want: 72 * time.Hour},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got, err := parseFlexibleDuration(tc.input)
			if err != nil {
				t.Fatalf("parseFlexibleDuration(%q) returned error: %v", tc.input, err)
			}
			if got != tc.want {
				t.Fatalf("parseFlexibleDuration(%q) = %s, want %s", tc.input, got, tc.want)
			}
		})
	}
}

func TestParseFlexibleDurationRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	inputs := []string{"", "0m", "10w", "minutes", "1.5h"}
	for _, input := range inputs {
		input := input
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			if _, err := parseFlexibleDuration(input); err == nil {
				t.Fatalf("parseFlexibleDuration(%q) unexpectedly succeeded", input)
			}
		})
	}
}

func TestBuildInteractiveArgsAddsNoAltScreen(t *testing.T) {
	t.Parallel()

	args := buildInteractiveArgs("/tmp/work", "prompt", "", false, false, true, launchOverrides{})
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "--no-alt-screen") {
		t.Fatalf("buildInteractiveArgs() did not include --no-alt-screen: %v", args)
	}
}

func TestBuildInteractiveArgsAddsLaunchOverrides(t *testing.T) {
	t.Parallel()

	args := buildInteractiveArgs("/tmp/work", "prompt", "", false, false, false, launchOverrides{
		Profile:              "safe-research",
		Model:                "gpt-5.3-codex-spark",
		ModelReasoningEffort: "high",
	})
	joined := strings.Join(args, " ")
	for _, expected := range []string{
		"--profile safe-research",
		"--model gpt-5.3-codex-spark",
		`--config model_reasoning_effort="high"`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("buildInteractiveArgs() missing %q: %v", expected, args)
		}
	}
}

func TestInteractiveLaunchBehaviorNewSession(t *testing.T) {
	t.Parallel()

	sendPromptOnLaunch, injectImmediately := interactiveLaunchBehavior("")
	if sendPromptOnLaunch {
		t.Fatal("interactiveLaunchBehavior() should not send the prompt on launch for a brand-new session")
	}
	if injectImmediately {
		t.Fatal("interactiveLaunchBehavior() should not inject immediately for a brand-new session")
	}
}

func TestInteractiveLaunchBehaviorResume(t *testing.T) {
	t.Parallel()

	sendPromptOnLaunch, injectImmediately := interactiveLaunchBehavior("session-123")
	if sendPromptOnLaunch {
		t.Fatal("interactiveLaunchBehavior() should not send the launch prompt when resuming")
	}
	if !injectImmediately {
		t.Fatal("interactiveLaunchBehavior() should inject immediately when resuming")
	}
}

func TestUseScreenIdleSchedulerDefaultsWhenIntervalUnset(t *testing.T) {
	t.Parallel()

	if !useScreenIdleScheduler(durationFlag{}, false) {
		t.Fatal("useScreenIdleScheduler() should default to screen-aware mode when no interval is set")
	}
}

func TestUseScreenIdleSchedulerUsesTimedModeWhenIntervalIsExplicit(t *testing.T) {
	t.Parallel()

	interval := durationFlag{
		value: 15 * time.Minute,
		set:   true,
	}
	if useScreenIdleScheduler(interval, false) {
		t.Fatal("useScreenIdleScheduler() should use timer mode when an explicit interval is set")
	}
	if !useScreenIdleScheduler(interval, true) {
		t.Fatal("useScreenIdleScheduler() should honor the explicit screen-aware flag")
	}
}

func TestRunHeartbeatModeDefaultsToScreenAwareSummary(t *testing.T) {
	t.Parallel()

	if got := runHeartbeatMode(durationFlag{}, false); got != screenIdleHeartbeatSummary() {
		t.Fatalf("runHeartbeatMode() = %q, want %q", got, screenIdleHeartbeatSummary())
	}
}

func TestRegisterRunFlagsOmitsSkipGitRepoCheck(t *testing.T) {
	t.Parallel()

	var opts sharedOptions
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	registerRunFlags(fs, &opts)
	if fs.Lookup("skip-git-repo-check") != nil {
		t.Fatal("registerRunFlags() unexpectedly exposed --skip-git-repo-check")
	}
	for _, flagName := range []string{"profile", "model", "model-reasoning-effort"} {
		if fs.Lookup(flagName) != nil {
			t.Fatalf("registerRunFlags() unexpectedly exposed --%s", flagName)
		}
	}
}

func TestMigrateLegacyProjectDirMovesLegacyState(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workdir := filepath.Join(root, "work")
	projectDir := filepath.Join(root, "runtime", "project")
	legacyDir := filepath.Join(workdir, ".codex-heartbeat")
	legacyLogsDir := filepath.Join(legacyDir, "logs")

	if err := os.MkdirAll(legacyLogsDir, 0o755); err != nil {
		t.Fatalf("mkdir legacy logs: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, "state.json"), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write legacy state: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legacyLogsDir, "2026-03-21.jsonl"), []byte("{}\n"), 0o644); err != nil {
		t.Fatalf("write legacy log: %v", err)
	}

	cfg := workspaceConfig{
		Workdir:    workdir,
		ProjectDir: projectDir,
		StatePath:  filepath.Join(projectDir, "state.json"),
		LockPath:   filepath.Join(projectDir, "heartbeat.lock"),
		LogsDir:    filepath.Join(projectDir, "logs"),
	}
	if err := migrateLegacyProjectDir(cfg); err != nil {
		t.Fatalf("migrateLegacyProjectDir() returned error: %v", err)
	}

	if _, err := os.Stat(legacyDir); !os.IsNotExist(err) {
		t.Fatalf("legacy runtime dir still exists after migration: %v", err)
	}
	if _, err := os.Stat(cfg.StatePath); err != nil {
		t.Fatalf("migrated state file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cfg.LogsDir, "2026-03-21.jsonl")); err != nil {
		t.Fatalf("migrated log file missing: %v", err)
	}
}

func TestStatusWorkspaceConfigUsesLegacyMigration(t *testing.T) {
	root := t.TempDir()

	previousHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if previousHome == "" {
			_ = os.Unsetenv("HOME")
			return
		}
		_ = os.Setenv("HOME", previousHome)
	})
	if err := os.Setenv("HOME", root); err != nil {
		t.Fatalf("set HOME: %v", err)
	}

	workdir := filepath.Join(root, "work")
	legacyDir := filepath.Join(workdir, ".codex-heartbeat")
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatalf("mkdir legacy dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legacyDir, "state.json"), []byte("{\"session_id\":\"abc\"}\n"), 0o644); err != nil {
		t.Fatalf("write legacy state: %v", err)
	}

	outputFile, err := os.CreateTemp(root, "status-output-*.json")
	if err != nil {
		t.Fatalf("create temp output file: %v", err)
	}
	outputPath := outputFile.Name()
	_ = outputFile.Close()

	previousStdout := os.Stdout
	outputWriter, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		t.Fatalf("open temp output file: %v", err)
	}
	os.Stdout = outputWriter
	t.Cleanup(func() {
		os.Stdout = previousStdout
	})

	if err := runStatusCommand([]string{"--workdir", workdir}); err != nil {
		t.Fatalf("runStatusCommand() returned error: %v", err)
	}
	_ = outputWriter.Close()
	os.Stdout = previousStdout

	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read status output: %v", err)
	}
	if !strings.Contains(string(output), "\"session_id\": \"abc\"") {
		t.Fatalf("status output missing migrated session id: %s", output)
	}
}

func TestStatusCommandIncludesProgramLaunchSettings(t *testing.T) {
	root := t.TempDir()

	previousHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if previousHome == "" {
			_ = os.Unsetenv("HOME")
			return
		}
		_ = os.Setenv("HOME", previousHome)
	})
	if err := os.Setenv("HOME", root); err != nil {
		t.Fatalf("set HOME: %v", err)
	}

	workdir := filepath.Join(root, "work")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}
	program := `# Program

Objective: show launch settings in status
Primary evaluator: go test ./cmd/codex-heartbeat
Profile: safe-research
Model: gpt-5.3-codex-spark
Model reasoning effort: high
`
	if err := os.WriteFile(filepath.Join(workdir, defaultProgramFilename), []byte(program), 0o644); err != nil {
		t.Fatalf("write program: %v", err)
	}

	outputFile, err := os.CreateTemp(root, "status-output-*.json")
	if err != nil {
		t.Fatalf("create temp output file: %v", err)
	}
	outputPath := outputFile.Name()
	_ = outputFile.Close()

	previousStdout := os.Stdout
	outputWriter, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		t.Fatalf("open temp output file: %v", err)
	}
	os.Stdout = outputWriter
	t.Cleanup(func() {
		os.Stdout = previousStdout
	})

	if err := runStatusCommand([]string{"--workdir", workdir}); err != nil {
		t.Fatalf("runStatusCommand() returned error: %v", err)
	}
	_ = outputWriter.Close()
	os.Stdout = previousStdout

	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read status output: %v", err)
	}
	for _, expected := range []string{
		"\"launch_settings\": {",
		"\"profile\": \"safe-research\"",
		"\"model\": \"gpt-5.3-codex-spark\"",
		"\"model_reasoning_effort\": \"high\"",
	} {
		if !strings.Contains(string(output), expected) {
			t.Fatalf("status output missing %q: %s", expected, output)
		}
	}
}

func TestStatusCommandIncludesHermesParityGap(t *testing.T) {
	root := t.TempDir()

	previousHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if previousHome == "" {
			_ = os.Unsetenv("HOME")
			return
		}
		_ = os.Setenv("HOME", previousHome)
	})
	if err := os.Setenv("HOME", root); err != nil {
		t.Fatalf("set HOME: %v", err)
	}

	workdir := filepath.Join(root, "work")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}

	outputFile, err := os.CreateTemp(root, "status-output-*.json")
	if err != nil {
		t.Fatalf("create temp output file: %v", err)
	}
	outputPath := outputFile.Name()
	_ = outputFile.Close()

	previousStdout := os.Stdout
	outputWriter, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		t.Fatalf("open temp output file: %v", err)
	}
	os.Stdout = outputWriter
	t.Cleanup(func() {
		os.Stdout = previousStdout
	})

	if err := runStatusCommand([]string{"--workdir", workdir}); err != nil {
		t.Fatalf("runStatusCommand() returned error: %v", err)
	}
	_ = outputWriter.Close()
	os.Stdout = previousStdout

	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read status output: %v", err)
	}
	for _, expected := range []string{
		"\"hermes_parity\": {",
		"\"equivalent\": false",
		"\"launch-time base/developer instruction control\"",
		"\"ephemeral prefill\"",
		"\"harmless canary scoring\"",
	} {
		if !strings.Contains(string(output), expected) {
			t.Fatalf("status output missing %q: %s", expected, output)
		}
	}
}

func TestRunInteractiveCommandNewIntervalLaunchIncludesPrompt(t *testing.T) {
	root := t.TempDir()
	workdir := filepath.Join(root, "work")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}

	promptText := "heartbeat prompt"
	promptPath := filepath.Join(root, "heartbeat.md")
	if err := os.WriteFile(promptPath, []byte(promptText+"\n"), 0o644); err != nil {
		t.Fatalf("write prompt file: %v", err)
	}

	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin dir: %v", err)
	}
	argsPath := filepath.Join(root, "codex-args.txt")
	scriptPath := filepath.Join(binDir, "codex")
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' \"$@\" > %q\nexit 0\n", argsPath)
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake codex: %v", err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TERM_PROGRAM", "")

	if err := runInteractiveCommand([]string{"--workdir", workdir, "--prompt", promptPath, "--interval", "15m"}); err != nil {
		t.Fatalf("runInteractiveCommand() returned error: %v", err)
	}

	argsData, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatalf("read fake codex args: %v", err)
	}
	args := strings.Split(strings.TrimSpace(string(argsData)), "\n")
	if len(args) == 0 {
		t.Fatal("fake codex did not receive any args")
	}
	if args[len(args)-1] == promptText {
		t.Fatalf("launch args unexpectedly included prompt %q; full args: %v", promptText, args)
	}
}

func TestRunInteractiveCommandScreenIdleLaunchIncludesPrompt(t *testing.T) {
	root := t.TempDir()
	workdir := filepath.Join(root, "work")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}

	promptText := "heartbeat prompt"
	promptPath := filepath.Join(root, "heartbeat.md")
	if err := os.WriteFile(promptPath, []byte(promptText+"\n"), 0o644); err != nil {
		t.Fatalf("write prompt file: %v", err)
	}

	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin dir: %v", err)
	}
	argsPath := filepath.Join(root, "codex-args.txt")
	scriptPath := filepath.Join(binDir, "codex")
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' \"$@\" > %q\nexit 0\n", argsPath)
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake codex: %v", err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TERM_PROGRAM", "")

	if err := runInteractiveCommand([]string{"--workdir", workdir, "--prompt", promptPath, "--screen-idle-heartbeat"}); err != nil {
		t.Fatalf("runInteractiveCommand() returned error: %v", err)
	}

	argsData, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatalf("read fake codex args: %v", err)
	}
	args := strings.Split(strings.TrimSpace(string(argsData)), "\n")
	if len(args) == 0 {
		t.Fatal("fake codex did not receive any args")
	}
	if args[len(args)-1] == promptText {
		t.Fatalf("launch args unexpectedly included prompt %q; full args: %v", promptText, args)
	}
}

func TestRunInteractiveCommandDefaultLaunchIncludesPrompt(t *testing.T) {
	root := t.TempDir()
	workdir := filepath.Join(root, "work")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}

	promptText := "heartbeat prompt"
	promptPath := filepath.Join(root, "heartbeat.md")
	if err := os.WriteFile(promptPath, []byte(promptText+"\n"), 0o644); err != nil {
		t.Fatalf("write prompt file: %v", err)
	}

	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin dir: %v", err)
	}
	argsPath := filepath.Join(root, "codex-args.txt")
	scriptPath := filepath.Join(binDir, "codex")
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' \"$@\" > %q\nexit 0\n", argsPath)
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake codex: %v", err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TERM_PROGRAM", "")

	if err := runInteractiveCommand([]string{"--workdir", workdir, "--prompt", promptPath}); err != nil {
		t.Fatalf("runInteractiveCommand() returned error: %v", err)
	}

	argsData, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatalf("read fake codex args: %v", err)
	}
	args := strings.Split(strings.TrimSpace(string(argsData)), "\n")
	if len(args) == 0 {
		t.Fatal("fake codex did not receive any args")
	}
	if args[len(args)-1] == promptText {
		t.Fatalf("launch args unexpectedly included prompt %q; full args: %v", promptText, args)
	}
}

func TestRunInteractiveCommandPassesLaunchOverrides(t *testing.T) {
	root := t.TempDir()
	workdir := filepath.Join(root, "work")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}

	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin dir: %v", err)
	}
	argsPath := filepath.Join(root, "codex-args.txt")
	scriptPath := filepath.Join(binDir, "codex")
	script := fmt.Sprintf("#!/bin/sh\nprintf '%%s\\n' \"$@\" > %q\nexit 0\n", argsPath)
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake codex: %v", err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TERM_PROGRAM", "")

	program := `# Program

Objective: safe launch overrides
Primary evaluator: go test ./cmd/codex-heartbeat
Profile: safe-research
Model: gpt-5.3-codex-spark
Model reasoning effort: high
`
	if err := os.WriteFile(filepath.Join(workdir, defaultProgramFilename), []byte(program), 0o644); err != nil {
		t.Fatalf("write program: %v", err)
	}

	if err := runInteractiveCommand([]string{"--workdir", workdir}); err != nil {
		t.Fatalf("runInteractiveCommand() returned error: %v", err)
	}

	argsData, err := os.ReadFile(argsPath)
	if err != nil {
		t.Fatalf("read fake codex args: %v", err)
	}
	joined := strings.Join(strings.Split(strings.TrimSpace(string(argsData)), "\n"), " ")
	for _, expected := range []string{
		"--profile safe-research",
		"--model gpt-5.3-codex-spark",
		`--config model_reasoning_effort="high"`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("runInteractiveCommand() missing %q in args: %s", expected, joined)
		}
	}
}

func TestInjectStartupPromptAfterDelay(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	root := t.TempDir()
	workdir := filepath.Join(root, "work")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}
	projectDir := filepath.Join(root, "runtime")
	cfg := workspaceConfig{
		LogsDir:    filepath.Join(projectDir, "logs"),
		ProjectDir: projectDir,
	}
	state := workspaceState{SessionID: "session-123"}
	promptTracker := newPromptInjectionTracker(time.Time{})
	promptPath := filepath.Join(root, "heartbeat.md")
	if err := os.WriteFile(promptPath, []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write prompt file: %v", err)
	}
	prompts, err := newPromptResolver(workdir, promptPath, projectDir, false)
	if err != nil {
		t.Fatalf("newPromptResolver() returned error: %v", err)
	}
	artifacts := newAutoresearchArtifacts(workdir, time.Now())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	injectStartupPromptAfterDelay(ctx, &output, prompts, artifacts, 10*time.Millisecond, promptTracker, cfg, &state, nil)

	got := output.String()
	if !strings.Contains(got, "hello") {
		t.Fatalf("injectStartupPromptAfterDelay() output = %q, want injected prompt text", got)
	}
	if promptTracker.LastPromptAt().IsZero() {
		t.Fatal("injectStartupPromptAfterDelay() should update the prompt tracker")
	}
}

func TestPromptSourceResolveCachesExplicitPrompt(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	promptPath := filepath.Join(projectDir, "heartbeat.md")
	if err := os.WriteFile(promptPath, []byte("first prompt\n"), 0o644); err != nil {
		t.Fatalf("write prompt file: %v", err)
	}

	prompts, err := newPromptSource(promptPath, projectDir)
	if err != nil {
		t.Fatalf("newPromptSource() returned error: %v", err)
	}

	if got, err := prompts.Resolve(); err != nil || got != "first prompt" {
		t.Fatalf("Resolve() = (%q, %v), want first prompt", got, err)
	}

	if err := os.WriteFile(promptPath, []byte("second prompt\n"), 0o644); err != nil {
		t.Fatalf("overwrite prompt file: %v", err)
	}
	if got, err := prompts.Resolve(); err != nil || got != "second prompt" {
		t.Fatalf("Resolve() after overwrite = (%q, %v), want second prompt", got, err)
	}

	if err := os.Remove(promptPath); err != nil {
		t.Fatalf("remove prompt file: %v", err)
	}
	if got, err := prompts.Resolve(); err != nil || got != "second prompt" {
		t.Fatalf("Resolve() with missing prompt file = (%q, %v), want cached second prompt", got, err)
	}
}

func TestPromptSourceResolveFailsWithoutPromptCache(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	promptPath := filepath.Join(projectDir, "missing.md")

	prompts, err := newPromptSource(promptPath, projectDir)
	if err != nil {
		t.Fatalf("newPromptSource() returned error: %v", err)
	}

	if _, err := prompts.Resolve(); err == nil {
		t.Fatal("Resolve() unexpectedly succeeded without a prompt file or cache")
	}
}

func TestRunInteractiveCommandRejectsConflictingHeartbeatFlags(t *testing.T) {
	t.Parallel()

	err := runInteractiveCommand([]string{"--workdir", t.TempDir(), "--interval", "15m", "--screen-idle-heartbeat"})
	if err == nil {
		t.Fatal("runInteractiveCommand() unexpectedly accepted conflicting heartbeat flags")
	}
	if !strings.Contains(err.Error(), "--interval and --screen-idle-heartbeat cannot be used together") {
		t.Fatalf("runInteractiveCommand() error = %v", err)
	}
}

func TestRunSubcommandHelpReturnsZero(t *testing.T) {
	if got := run([]string{"run", "--help"}); got != 0 {
		t.Fatalf("run(run --help) = %d, want 0", got)
	}
}

func TestRootHelpReturnsZero(t *testing.T) {
	if got := run([]string{"--help"}); got != 0 {
		t.Fatalf("run(--help) = %d, want 0", got)
	}
}

func TestStatusSubcommandHelpReturnsZero(t *testing.T) {
	if got := run([]string{"status", "--help"}); got != 0 {
		t.Fatalf("run(status --help) = %d, want 0", got)
	}
}

func TestRootFlagsDefaultToRun(t *testing.T) {
	root := t.TempDir()
	workdir := filepath.Join(root, "work")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}

	promptPath := filepath.Join(root, "heartbeat.md")
	if err := os.WriteFile(promptPath, []byte("heartbeat prompt\n"), 0o644); err != nil {
		t.Fatalf("write prompt file: %v", err)
	}

	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir bin dir: %v", err)
	}
	scriptPath := filepath.Join(binDir, "codex")
	script := "#!/bin/sh\nexit 0\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake codex: %v", err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("TERM_PROGRAM", "")

	if got := run([]string{"--workdir", workdir, "--prompt", promptPath, "--interval", "15m"}); got != 0 {
		t.Fatalf("run(default flags form) = %d, want 0", got)
	}
}

func TestResolveNoAltScreenRejectsConflictingFlags(t *testing.T) {
	t.Parallel()

	if _, err := resolveNoAltScreen(true, true); err == nil {
		t.Fatal("resolveNoAltScreen() unexpectedly accepted conflicting flags")
	}
}

func TestResolveNoAltScreenGhosttyDefault(t *testing.T) {
	t.Parallel()

	previous := os.Getenv("TERM_PROGRAM")
	t.Cleanup(func() {
		if previous == "" {
			_ = os.Unsetenv("TERM_PROGRAM")
			return
		}
		_ = os.Setenv("TERM_PROGRAM", previous)
	})

	_ = os.Setenv("TERM_PROGRAM", "ghostty")
	got, err := resolveNoAltScreen(false, false)
	if err != nil {
		t.Fatalf("resolveNoAltScreen() returned error: %v", err)
	}

	want := runtime.GOOS == "darwin"
	if got != want {
		t.Fatalf("resolveNoAltScreen() = %v, want %v", got, want)
	}
}

func TestRunTerminalTitleIncludesWorkspace(t *testing.T) {
	t.Parallel()

	got := runTerminalTitle(workspaceConfig{Workdir: "/tmp/codex-heartbeat"})
	if !strings.Contains(got, "codex-heartbeat | ") {
		t.Fatalf("runTerminalTitle() missing heartbeat prefix: %q", got)
	}
	if !strings.Contains(got, "/tmp/codex-heartbeat") {
		t.Fatalf("runTerminalTitle() missing workspace: %q", got)
	}
}

func TestTerminalTitleSequenceSanitizesControlCharacters(t *testing.T) {
	t.Parallel()

	got := terminalTitleSequence("heartbeat\tone\n two\a")
	want := "\033]0;heartbeat one two\007"
	if got != want {
		t.Fatalf("terminalTitleSequence() = %q, want %q", got, want)
	}
}
