package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScreenFindSessionRolloutPathHonorsCodexHome(t *testing.T) {
	codexHome := t.TempDir()
	t.Setenv("CODEX_HOME", codexHome)

	sessionID := "019d3608-4633-7260-9f2b-cf68b279f592"
	wantPath := writeScreenRolloutFixture(t, codexHome, sessionID, strings.Join([]string{
		`{"timestamp":"2026-03-28T22:53:10.170Z","type":"session_meta","payload":{"id":"SESSION_ID","cwd":"/tmp/work","timestamp":"2026-03-28T22:53:10.170Z"}}`,
		`{"timestamp":"2026-03-28T22:53:10.171Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"turn-1"}}`,
	}, "\n")+"\n")

	got, err := findSessionRolloutPath(sessionID)
	if err != nil {
		t.Fatalf("findSessionRolloutPath() returned error: %v", err)
	}
	if got != wantPath {
		t.Fatalf("findSessionRolloutPath() = %q, want %q", got, wantPath)
	}
}

func TestScreenClassifySessionRolloutPathTaskCompleteIsIdle(t *testing.T) {
	t.Parallel()

	path := writeScreenRolloutFixture(t, t.TempDir(), "session-idle", strings.Join([]string{
		`{"timestamp":"2026-03-28T22:53:10.170Z","type":"session_meta","payload":{"id":"SESSION_ID","cwd":"/tmp/work","timestamp":"2026-03-28T22:53:10.170Z"}}`,
		`{"timestamp":"2026-03-28T22:53:10.171Z","type":"event_msg","payload":{"type":"task_complete","turn_id":"turn-1"}}`,
	}, "\n")+"\n")

	got, reason, err := classifySessionRolloutPath(path)
	if err != nil {
		t.Fatalf("classifySessionRolloutPath() returned error: %v", err)
	}
	if got != screenStateIdle || reason != "rollout_task_complete" {
		t.Fatalf("classifySessionRolloutPath() = (%v, %q), want (%v, %q)", got, reason, screenStateIdle, "rollout_task_complete")
	}
}

func TestScreenClassifySessionRolloutPathTurnStartedIsWorking(t *testing.T) {
	t.Parallel()

	path := writeScreenRolloutFixture(t, t.TempDir(), "session-working", strings.Join([]string{
		`{"timestamp":"2026-03-28T22:53:10.170Z","type":"session_meta","payload":{"id":"SESSION_ID","cwd":"/tmp/work","timestamp":"2026-03-28T22:53:10.170Z"}}`,
		`{"timestamp":"2026-03-28T22:53:10.171Z","type":"event_msg","payload":{"type":"task_started","turn_id":"turn-1"}}`,
	}, "\n")+"\n")

	got, reason, err := classifySessionRolloutPath(path)
	if err != nil {
		t.Fatalf("classifySessionRolloutPath() returned error: %v", err)
	}
	if got != screenStateWorking || reason != "rollout_task_started" {
		t.Fatalf("classifySessionRolloutPath() = (%v, %q), want (%v, %q)", got, reason, screenStateWorking, "rollout_task_started")
	}
}

func TestScreenClassifySessionRolloutPathPendingFunctionCallIsWorking(t *testing.T) {
	t.Parallel()

	path := writeScreenRolloutFixture(t, t.TempDir(), "session-pending", strings.Join([]string{
		`{"timestamp":"2026-03-28T22:53:10.170Z","type":"session_meta","payload":{"id":"SESSION_ID","cwd":"/tmp/work","timestamp":"2026-03-28T22:53:10.170Z"}}`,
		`{"timestamp":"2026-03-28T22:53:10.171Z","type":"response_item","payload":{"type":"function_call","call_id":"call-123","name":"exec_command","arguments":"{}"}}`,
	}, "\n")+"\n")

	got, reason, err := classifySessionRolloutPath(path)
	if err != nil {
		t.Fatalf("classifySessionRolloutPath() returned error: %v", err)
	}
	if got != screenStateWorking || reason != "rollout_pending_function_call" {
		t.Fatalf("classifySessionRolloutPath() = (%v, %q), want (%v, %q)", got, reason, screenStateWorking, "rollout_pending_function_call")
	}
}

func TestScreenRolloutInspectorResolvesAmbiguousScreen(t *testing.T) {
	codexHome := t.TempDir()
	t.Setenv("CODEX_HOME", codexHome)

	sessionID := "session-ambiguous-idle"
	writeScreenRolloutFixture(t, codexHome, sessionID, strings.Join([]string{
		`{"timestamp":"2026-03-28T22:53:10.170Z","type":"session_meta","payload":{"id":"SESSION_ID","cwd":"/tmp/work","timestamp":"2026-03-28T22:53:10.170Z"}}`,
		`{"timestamp":"2026-03-28T22:53:10.171Z","type":"event_msg","payload":{"type":"context_compacted"}}`,
	}, "\n")+"\n")

	inspector := newSessionRolloutInspector()
	got, reason := inspector.Resolve(screenStateAmbiguous, sessionID)
	if got != screenStateIdle || reason != "rollout_context_compacted" {
		t.Fatalf("Resolve() = (%v, %q), want (%v, %q)", got, reason, screenStateIdle, "rollout_context_compacted")
	}
}

func TestScreenRolloutInspectorOverridesIdleScreenWhenRolloutIsWorking(t *testing.T) {
	codexHome := t.TempDir()
	t.Setenv("CODEX_HOME", codexHome)

	sessionID := "session-idle-screen-working-rollout"
	writeScreenRolloutFixture(t, codexHome, sessionID, strings.Join([]string{
		`{"timestamp":"2026-03-28T22:53:10.170Z","type":"session_meta","payload":{"id":"SESSION_ID","cwd":"/tmp/work","timestamp":"2026-03-28T22:53:10.170Z"}}`,
		`{"timestamp":"2026-03-28T22:53:10.171Z","type":"event_msg","payload":{"type":"task_started"}}`,
	}, "\n")+"\n")

	inspector := newSessionRolloutInspector()
	got, reason := inspector.Resolve(screenStateIdle, sessionID)
	if got != screenStateWorking || reason != "rollout_task_started" {
		t.Fatalf("Resolve() = (%v, %q), want (%v, %q)", got, reason, screenStateWorking, "rollout_task_started")
	}
}

func writeScreenRolloutFixture(t *testing.T, codexHome, sessionID, contents string) string {
	t.Helper()

	path := filepath.Join(codexHome, "sessions", "2026", "03", "28", "rollout-2026-03-28T22-53-10-"+sessionID+".jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) returned error: %v", filepath.Dir(path), err)
	}

	contents = strings.ReplaceAll(contents, "SESSION_ID", sessionID)
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) returned error: %v", path, err)
	}
	return path
}
