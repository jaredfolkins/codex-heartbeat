package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func loadScreenFixture(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join("testdata", "screen", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) returned error: %v", path, err)
	}
	return string(data)
}

func TestClassifyScreenSnapshotFixtures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		file string
		want screenState
	}{
		{name: "active working", file: "active_working.txt", want: screenStateWorking},
		{name: "active without interrupt", file: "active_no_interrupt.txt", want: screenStateWorking},
		{name: "active wrapped waiting", file: "active_wrapped_waiting.txt", want: screenStateWorking},
		{name: "idle token usage", file: "idle_token_usage.txt", want: screenStateIdle},
		{name: "pending steer", file: "pending_steer.txt", want: screenStateIdle},
		{name: "history background", file: "history_background_terminal.txt", want: screenStateIdle},
		{name: "footer background", file: "footer_background_terminal.txt", want: screenStateIdle},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := classifyScreenSnapshot(loadScreenFixture(t, tc.file)); got != tc.want {
				t.Fatalf("classifyScreenSnapshot(%s) = %v, want %v", tc.file, got, tc.want)
			}
		})
	}
}

func TestClassifyScreenSnapshotFixtureMutations(t *testing.T) {
	t.Parallel()

	active := loadScreenFixture(t, "active_working.txt")
	activeMutations := []string{
		strings.ReplaceAll(active, "•", "◦"),
		strings.ReplaceAll(active, "Working", "WORKING"),
		strings.ReplaceAll(active, "esc to interrupt", "ESC TO INTERRUPT"),
	}
	for idx, snapshot := range activeMutations {
		if got := classifyScreenSnapshot(snapshot); got != screenStateWorking {
			t.Fatalf("classifyScreenSnapshot(active mutation %d) = %v, want working", idx, got)
		}
	}

	pendingSteer := loadScreenFixture(t, "pending_steer.txt")
	pendingMutations := []string{
		strings.ReplaceAll(pendingSteer, "press esc to interrupt and send immediately", "PRESS ESC TO INTERRUPT AND SEND IMMEDIATELY"),
		strings.ReplaceAll(pendingSteer, "Messages to be submitted after next tool call", "messages to be submitted after next tool call"),
	}
	for idx, snapshot := range pendingMutations {
		if got := classifyScreenSnapshot(snapshot); got != screenStateIdle {
			t.Fatalf("classifyScreenSnapshot(pending mutation %d) = %v, want idle", idx, got)
		}
	}
}

func TestEvaluateScreenIdlePollSequence(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.March, 28, 21, 0, 0, 0, time.UTC)
	lastPromptAt := now.Add(-30 * time.Minute)
	idlePolls := 0

	steps := []struct {
		name       string
		advance    time.Duration
		quiet      bool
		state      screenState
		wantPolls  int
		wantInject bool
		wantReason string
	}{
		{name: "working resets", advance: 0, quiet: true, state: screenStateWorking, wantPolls: 0, wantReason: "screen_working"},
		{name: "idle begins", advance: screenIdlePollInterval, quiet: true, state: screenStateIdle, wantPolls: 1, wantReason: "idle_accumulating"},
		{name: "recent input delays fire", advance: screenIdlePollInterval, quiet: false, state: screenStateIdle, wantPolls: 2, wantReason: "idle_accumulating_recent_input"},
		{name: "ready but still recent input", advance: screenIdlePollInterval, quiet: false, state: screenStateIdle, wantPolls: 3, wantReason: "idle_ready_recent_input"},
		{name: "quiet idle injects", advance: screenIdlePollInterval, quiet: true, state: screenStateIdle, wantInject: true, wantReason: "idle_threshold_reached"},
	}

	current := now
	for _, step := range steps {
		current = current.Add(step.advance)
		decision := evaluateScreenIdlePoll(current, step.quiet, step.state, idlePolls, lastPromptAt)
		if decision.nextIdlePolls != step.wantPolls || decision.shouldInject != step.wantInject || decision.reason != step.wantReason {
			t.Fatalf("%s: evaluateScreenIdlePoll() = (%d, %t, %q), want (%d, %t, %q)", step.name, decision.nextIdlePolls, decision.shouldInject, decision.reason, step.wantPolls, step.wantInject, step.wantReason)
		}
		idlePolls = decision.nextIdlePolls
	}
}

func TestHeartbeatReplaySequence(t *testing.T) {
	t.Parallel()

	lastPromptAt := time.Date(2026, time.March, 28, 21, 30, 0, 0, time.UTC).Add(-30 * time.Minute)
	now := time.Date(2026, time.March, 28, 21, 30, 0, 0, time.UTC)
	idlePolls := 0

	steps := []struct {
		name       string
		file       string
		quiet      bool
		wantState  screenState
		wantPolls  int
		wantInject bool
		wantReason string
	}{
		{name: "active status row", file: "active_working.txt", quiet: true, wantState: screenStateWorking, wantPolls: 0, wantReason: "screen_working"},
		{name: "active without interrupt hint", file: "active_no_interrupt.txt", quiet: true, wantState: screenStateWorking, wantPolls: 0, wantReason: "screen_working"},
		{name: "footer noise does not force working", file: "footer_background_terminal.txt", quiet: true, wantState: screenStateIdle, wantPolls: 1, wantReason: "idle_accumulating"},
		{name: "idle token usage accumulates", file: "idle_token_usage.txt", quiet: true, wantState: screenStateIdle, wantPolls: 2, wantReason: "idle_accumulating"},
		{name: "third quiet idle injects", file: "history_background_terminal.txt", quiet: true, wantState: screenStateIdle, wantInject: true, wantReason: "idle_threshold_reached"},
	}

	for idx, step := range steps {
		now = now.Add(screenIdlePollInterval)
		state := classifyScreenSnapshot(loadScreenFixture(t, step.file))
		if state != step.wantState {
			t.Fatalf("%s: classifyScreenSnapshot(%s) = %v, want %v", step.name, step.file, state, step.wantState)
		}

		decision := evaluateScreenIdlePoll(now, step.quiet, state, idlePolls, lastPromptAt)
		if decision.nextIdlePolls != step.wantPolls || decision.shouldInject != step.wantInject || decision.reason != step.wantReason {
			t.Fatalf("%s (step %d): evaluateScreenIdlePoll() = (%d, %t, %q), want (%d, %t, %q)", step.name, idx, decision.nextIdlePolls, decision.shouldInject, decision.reason, step.wantPolls, step.wantInject, step.wantReason)
		}
		idlePolls = decision.nextIdlePolls
	}
}
