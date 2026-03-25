package main

import (
	"io"
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

	if got, want := screenIdleHeartbeatSummary(), "screen-idle=3x10s/fallback=60m"; got != want {
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

func TestAdvanceScreenIdlePolls(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		idlePolls  int
		quiet      bool
		state      screenState
		wantPolls  int
		wantInject bool
	}{
		{name: "first idle poll", idlePolls: 0, quiet: true, state: screenStateIdle, wantPolls: 1},
		{name: "second idle poll", idlePolls: 1, quiet: true, state: screenStateIdle, wantPolls: 2},
		{name: "third idle poll injects", idlePolls: 2, quiet: true, state: screenStateIdle, wantInject: true},
		{name: "recent input resets idle accumulation", idlePolls: 2, quiet: false, state: screenStateIdle},
		{name: "working screen resets idle accumulation", idlePolls: 2, quiet: true, state: screenStateWorking},
		{name: "ambiguous screen resets idle accumulation", idlePolls: 2, quiet: true, state: screenStateAmbiguous},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			gotPolls, gotInject := advanceScreenIdlePolls(tc.idlePolls, tc.quiet, tc.state)
			if gotPolls != tc.wantPolls || gotInject != tc.wantInject {
				t.Fatalf("advanceScreenIdlePolls(%d, %t, %v) = (%d, %t), want (%d, %t)", tc.idlePolls, tc.quiet, tc.state, gotPolls, gotInject, tc.wantPolls, tc.wantInject)
			}
		})
	}
}
