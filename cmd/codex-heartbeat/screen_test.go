package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestClassifyScreenSnapshotWorking(t *testing.T) {
	t.Parallel()

	snapshot := "\u203a Continue task\n\u2022 Working (3m 02s \u2022 esc to interrupt)\n"
	if got := classifyScreenSnapshot(snapshot); got != screenStateWorking {
		t.Fatalf("classifyScreenSnapshot() = %v, want working", got)
	}
}

func TestClassifyScreenSnapshotIdle(t *testing.T) {
	t.Parallel()

	snapshot := "\u203a Continue task\nToken usage: total=20 input=10 output=10\n"
	if got := classifyScreenSnapshot(snapshot); got != screenStateIdle {
		t.Fatalf("classifyScreenSnapshot() = %v, want idle", got)
	}
}

func TestClassifyScreenSnapshotAmbiguous(t *testing.T) {
	t.Parallel()

	if got := classifyScreenSnapshot("OpenAI Codex loading /model to change"); got != screenStateAmbiguous {
		t.Fatalf("classifyScreenSnapshot() = %v, want ambiguous", got)
	}
}

func TestTerminalScreenSnapshotTracksCurrentStatus(t *testing.T) {
	t.Parallel()

	screen := newTerminalScreen(80, 10)
	active := "\x1b[9;1HWorking (3m 02s \u2022 esc to interrupt)\x1b[10;1H\u203a Continue task"
	if _, err := screen.Write([]byte(active)); err != nil {
		t.Fatalf("Write(active) returned error: %v", err)
	}
	if got := classifyScreenSnapshot(screen.Snapshot()); got != screenStateWorking {
		t.Fatalf("classifyScreenSnapshot(active snapshot) = %v, want working", got)
	}

	idle := "\x1b[9;1H\x1b[2K\x1b[10;1H\u203a Continue task\x1b[11;1HToken usage: total=20"
	if _, err := screen.Write([]byte(idle)); err != nil {
		t.Fatalf("Write(idle) returned error: %v", err)
	}
	if got := classifyScreenSnapshot(screen.Snapshot()); got != screenStateIdle {
		t.Fatalf("classifyScreenSnapshot(idle snapshot) = %v, want idle", got)
	}
}

func TestTerminalScreenRecentSnapshotIgnoresStaleWorkingText(t *testing.T) {
	t.Parallel()

	screen := newTerminalScreen(80, 12)
	screen.row = 11
	screen.cells[0] = []rune("Working (3m 02s • esc to interrupt)" + strings.Repeat(" ", 45))
	screen.cells[10] = []rune("› Continue task" + strings.Repeat(" ", 65))
	screen.cells[11] = []rune("Token usage: total=20" + strings.Repeat(" ", 59))

	if got := classifyScreenSnapshot(screen.Snapshot()); got != screenStateWorking {
		t.Fatalf("classifyScreenSnapshot(full snapshot) = %v, want working", got)
	}
	if got := classifyScreenSnapshot(screen.RecentSnapshot(screenIdleRecentLines)); got != screenStateIdle {
		t.Fatalf("classifyScreenSnapshot(recent snapshot) = %v, want idle", got)
	}
}

func TestTerminalScreenRecentSnapshotKeepsLiveWorkingText(t *testing.T) {
	t.Parallel()

	screen := newTerminalScreen(80, 12)
	screen.row = 11
	screen.cells[10] = []rune("› Continue task" + strings.Repeat(" ", 65))
	screen.cells[11] = []rune("Working (3m 02s • esc to interrupt)" + strings.Repeat(" ", 45))

	if got := classifyScreenSnapshot(screen.RecentSnapshot(screenIdleRecentLines)); got != screenStateWorking {
		t.Fatalf("classifyScreenSnapshot(recent working snapshot) = %v, want working", got)
	}
}

func TestScreenIdleHeartbeatSummary(t *testing.T) {
	t.Parallel()

	if got, want := screenIdleHeartbeatSummary(), "screen-idle=3x5s/quiet=20s/fallback=60m"; got != want {
		t.Fatalf("screenIdleHeartbeatSummary() = %q, want %q", got, want)
	}
}

func TestScreenIdleFallbackDue(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 25, 12, 0, 0, 0, time.UTC)
	lastPromptAt := now.Add(-screenIdleFallbackWait)

	if !screenIdleFallbackDue(now, lastPromptAt, true) {
		t.Fatal("screenIdleFallbackDue() should fire when the fallback window elapses")
	}
	if screenIdleFallbackDue(now, lastPromptAt.Add(time.Second), true) {
		t.Fatal("screenIdleFallbackDue() should not fire before the fallback window elapses")
	}
	if screenIdleFallbackDue(now, lastPromptAt, false) {
		t.Fatal("screenIdleFallbackDue() should respect the quiet-input gate")
	}
}

func TestUserInputTrackerQuietWindow(t *testing.T) {
	t.Parallel()

	tracker := &userInputTracker{}
	now := time.Date(2026, time.March, 25, 12, 0, 0, 0, time.UTC)

	if !tracker.IsQuiet(now, screenIdleQuietWindow) {
		t.Fatal("tracker without input should be quiet")
	}

	tracker.Mark(now)

	if tracker.IsQuiet(now.Add(5*time.Second), screenIdleQuietWindow) {
		t.Fatal("tracker with recent input should not be quiet")
	}
	if !tracker.IsQuiet(now.Add(screenIdleQuietWindow), screenIdleQuietWindow) {
		t.Fatal("tracker should become quiet once the quiet window passes")
	}
}

func TestTrackUserInputMarksActivity(t *testing.T) {
	t.Parallel()

	tracker := &userInputTracker{}
	reader := trackUserInput(strings.NewReader("hello"), tracker)

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll(trackUserInput()) returned error: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("ReadAll(trackUserInput()) = %q, want %q", data, "hello")
	}
	if tracker.IsQuiet(time.Now(), time.Hour) {
		t.Fatal("tracker should record recent input activity")
	}
}

func TestEvaluateScreenIdlePoll(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 25, 12, 0, 0, 0, time.UTC)
	lastPromptAt := now.Add(-30 * time.Minute)

	tests := []struct {
		name       string
		idlePolls  int
		quiet      bool
		state      screenState
		lastPrompt time.Time
		wantPolls  int
		wantInject bool
		wantReason string
	}{
		{name: "first idle poll", idlePolls: 0, quiet: true, state: screenStateIdle, lastPrompt: lastPromptAt, wantPolls: 1, wantReason: "idle_accumulating"},
		{name: "second idle poll", idlePolls: 1, quiet: true, state: screenStateIdle, lastPrompt: lastPromptAt, wantPolls: 2, wantReason: "idle_accumulating"},
		{name: "third idle poll injects", idlePolls: 2, quiet: true, state: screenStateIdle, lastPrompt: lastPromptAt, wantInject: true, wantReason: "idle_threshold_reached"},
		{name: "recent input still accumulates idle evidence", idlePolls: 1, quiet: false, state: screenStateIdle, lastPrompt: lastPromptAt, wantPolls: 2, wantReason: "idle_accumulating_recent_input"},
		{name: "recent input holds ready idle screen", idlePolls: 2, quiet: false, state: screenStateIdle, lastPrompt: lastPromptAt, wantPolls: 3, wantReason: "idle_ready_recent_input"},
		{name: "working screen resets idle accumulation", idlePolls: 2, quiet: true, state: screenStateWorking, lastPrompt: lastPromptAt, wantReason: "screen_working"},
		{name: "ambiguous screen resets idle accumulation", idlePolls: 2, quiet: true, state: screenStateAmbiguous, lastPrompt: lastPromptAt, wantReason: "screen_ambiguous"},
		{name: "fallback injects after 60m", idlePolls: 0, quiet: true, state: screenStateWorking, lastPrompt: now.Add(-screenIdleFallbackWait), wantInject: true, wantReason: "fallback_due"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			decision := evaluateScreenIdlePoll(now, tc.quiet, tc.state, tc.idlePolls, tc.lastPrompt)
			if decision.nextIdlePolls != tc.wantPolls || decision.shouldInject != tc.wantInject || decision.reason != tc.wantReason {
				t.Fatalf("evaluateScreenIdlePoll() = (%d, %t, %q), want (%d, %t, %q)", decision.nextIdlePolls, decision.shouldInject, decision.reason, tc.wantPolls, tc.wantInject, tc.wantReason)
			}
		})
	}
}

func TestPersistScreenDiagnostics(t *testing.T) {
	t.Parallel()

	projectDir := t.TempDir()
	logsDir := filepath.Join(projectDir, "logs")
	cfg := workspaceConfig{
		ProjectDir: projectDir,
		LogsDir:    logsDir,
	}

	now := time.Date(2026, time.March, 25, 12, 0, 0, 0, time.UTC)
	runtimeState := screenRuntimeState{
		SessionID:     "session-123",
		Scheduler:     screenIdleHeartbeatSummary(),
		ScreenState:   "idle",
		Quiet:         true,
		IdlePolls:     2,
		Reason:        "idle_accumulating",
		LastCheckedAt: now,
		LastPromptAt:  now.Add(-time.Minute),
		Snapshot:      "Token usage: total=20",
	}
	poll := screenPollRecord{
		Timestamp:    now.Format(time.RFC3339),
		SessionID:    "session-123",
		Scheduler:    screenIdleHeartbeatSummary(),
		ScreenState:  "idle",
		Quiet:        true,
		IdlePolls:    2,
		Reason:       "idle_accumulating",
		LastPromptAt: now.Add(-time.Minute).Format(time.RFC3339),
		Snapshot:     "Token usage: total=20",
	}

	persistScreenDiagnostics(cfg, runtimeState, poll)

	statePath := screenStateFilePath(projectDir)
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("ReadFile(screen state) returned error: %v", err)
	}

	var storedState screenRuntimeState
	if err := json.Unmarshal(data, &storedState); err != nil {
		t.Fatalf("Unmarshal(screen state) returned error: %v", err)
	}
	if storedState.ScreenState != "idle" || storedState.Reason != "idle_accumulating" {
		t.Fatalf("stored screen state = %+v", storedState)
	}

	logPath := filepath.Join(logsDir, now.Format("2006-01-02")+"-screen.jsonl")
	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("ReadFile(screen log) returned error: %v", err)
	}
	if !strings.Contains(string(logData), "\"reason\":\"idle_accumulating\"") {
		t.Fatalf("screen log missing expected reason: %s", logData)
	}
}
