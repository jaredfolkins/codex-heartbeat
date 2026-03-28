package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPromptResolverPrefersExplicitPromptOverProgram(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workdir := filepath.Join(root, "work")
	projectDir := filepath.Join(root, "runtime")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(workdir, defaultProgramFilename), []byte("# Program\n\nObjective: from program\nPrimary evaluator: go test ./...\n"), 0o644); err != nil {
		t.Fatalf("write program: %v", err)
	}
	promptPath := filepath.Join(root, "override.md")
	if err := os.WriteFile(promptPath, []byte("explicit override\n"), 0o644); err != nil {
		t.Fatalf("write explicit prompt: %v", err)
	}

	resolver, err := newPromptResolver(workdir, promptPath, projectDir)
	if err != nil {
		t.Fatalf("newPromptResolver() returned error: %v", err)
	}

	artifacts := newAutoresearchArtifacts(workdir, time.Date(2026, time.March, 28, 20, 0, 0, 0, time.UTC))
	resolution, err := resolver.Resolve(artifacts)
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	if resolution.Source != promptSourceCLI {
		t.Fatalf("Resolve().Source = %q, want %q", resolution.Source, promptSourceCLI)
	}
	if resolution.Text != "explicit override" {
		t.Fatalf("Resolve().Text = %q, want explicit override", resolution.Text)
	}
}

func TestPromptResolverUsesProgramPromptByDefault(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workdir := filepath.Join(root, "work")
	projectDir := filepath.Join(root, "runtime")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}

	program := `# Program

Objective: tighten the retry loop
Primary evaluator: go test ./cmd/codex-heartbeat
Prompt mode: autoresearch
Council after failures: 2
Checkpoint commits: true

## Notes

- Keep the hypothesis narrow.
`
	if err := os.WriteFile(filepath.Join(workdir, defaultProgramFilename), []byte(program), 0o644); err != nil {
		t.Fatalf("write program: %v", err)
	}

	resolver, err := newPromptResolver(workdir, "", projectDir)
	if err != nil {
		t.Fatalf("newPromptResolver() returned error: %v", err)
	}

	artifacts := newAutoresearchArtifacts(workdir, time.Date(2026, time.March, 28, 20, 1, 0, 0, time.UTC))
	resolution, err := resolver.Resolve(artifacts)
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	if resolution.Source != promptSourceProgram {
		t.Fatalf("Resolve().Source = %q, want %q", resolution.Source, promptSourceProgram)
	}
	if !strings.Contains(resolution.Text, "tighten the retry loop") {
		t.Fatalf("Resolve().Text missing objective: %q", resolution.Text)
	}
	if !strings.Contains(resolution.Text, "go test ./cmd/codex-heartbeat") {
		t.Fatalf("Resolve().Text missing evaluator: %q", resolution.Text)
	}
	if !strings.Contains(resolution.Text, "Do not start with the 3-agent council") {
		t.Fatalf("Resolve().Text missing fallback council policy: %q", resolution.Text)
	}
}

func TestPromptResolverFallsBackToEmbeddedTemplate(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workdir := filepath.Join(root, "work")
	projectDir := filepath.Join(root, "runtime")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}

	resolver, err := newPromptResolver(workdir, "", projectDir)
	if err != nil {
		t.Fatalf("newPromptResolver() returned error: %v", err)
	}

	artifacts := newAutoresearchArtifacts(workdir, time.Date(2026, time.March, 28, 20, 2, 0, 0, time.UTC))
	resolution, err := resolver.Resolve(artifacts)
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	if resolution.Source != promptSourceEmbedded {
		t.Fatalf("Resolve().Source = %q, want %q", resolution.Source, promptSourceEmbedded)
	}
	if !strings.Contains(resolution.Text, "Autoresearch loop contract") {
		t.Fatalf("Resolve().Text missing fallback template: %q", resolution.Text)
	}
}

func TestPromptResolverManualTestFirstMode(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workdir := filepath.Join(root, "work")
	projectDir := filepath.Join(root, "runtime")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}

	program := `# Program

Objective: prepare a manual fix
Primary evaluator: make manual-lab-up
Prompt mode: manual-test-first

## Notes

- Stop before the final human gate.
`
	if err := os.WriteFile(filepath.Join(workdir, defaultProgramFilename), []byte(program), 0o644); err != nil {
		t.Fatalf("write program: %v", err)
	}

	resolver, err := newPromptResolver(workdir, "", projectDir)
	if err != nil {
		t.Fatalf("newPromptResolver() returned error: %v", err)
	}

	artifacts := newAutoresearchArtifacts(workdir, time.Date(2026, time.March, 28, 20, 3, 0, 0, time.UTC))
	resolution, err := resolver.Resolve(artifacts)
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	if !strings.Contains(resolution.Text, "stop before the final human gate") {
		t.Fatalf("Resolve().Text missing manual gate guidance: %q", resolution.Text)
	}
}

func TestPromptResolverProgramTakesPrecedenceWhenExplicitPromptFlagIsUnset(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workdir := filepath.Join(root, "work")
	projectDir := filepath.Join(root, "runtime")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}

	program := `# Program

Objective: use the human program
Primary evaluator: go test ./...
`
	if err := os.WriteFile(filepath.Join(workdir, defaultProgramFilename), []byte(program), 0o644); err != nil {
		t.Fatalf("write program: %v", err)
	}

	explicitPromptPath := filepath.Join(root, "explicit.md")
	if err := os.WriteFile(explicitPromptPath, []byte("stale explicit prompt\n"), 0o644); err != nil {
		t.Fatalf("write explicit prompt: %v", err)
	}
	explicitPrompt, err := newPromptSource(explicitPromptPath, projectDir)
	if err != nil {
		t.Fatalf("newPromptSource() returned error: %v", err)
	}
	if _, err := explicitPrompt.Resolve(); err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}
	if err := os.Remove(explicitPromptPath); err != nil {
		t.Fatalf("remove explicit prompt: %v", err)
	}

	resolver, err := newPromptResolver(workdir, "", projectDir)
	if err != nil {
		t.Fatalf("newPromptResolver() returned error: %v", err)
	}
	artifacts := newAutoresearchArtifacts(workdir, time.Date(2026, time.March, 28, 20, 4, 0, 0, time.UTC))
	resolution, err := resolver.Resolve(artifacts)
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	if resolution.Source != promptSourceProgram {
		t.Fatalf("Resolve().Source = %q, want %q", resolution.Source, promptSourceProgram)
	}
	if strings.Contains(resolution.Text, "stale explicit prompt") {
		t.Fatalf("Resolve().Text incorrectly reused cached explicit prompt: %q", resolution.Text)
	}
}

func TestLoadPriorInsightsNewestFirst(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	targetDir := filepath.Join(root, targetDirName)
	olderDir := filepath.Join(targetDir, "run-20260327T120000Z")
	newerDir := filepath.Join(targetDir, "run-20260328T120000Z")
	currentDir := filepath.Join(targetDir, "run-20260329T120000Z")
	for _, dir := range []string{olderDir, newerDir, currentDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	if err := os.WriteFile(filepath.Join(olderDir, "insights.md"), []byte("# Insights\n\n- older insight\n"), 0o644); err != nil {
		t.Fatalf("write older insights: %v", err)
	}
	if err := os.WriteFile(filepath.Join(newerDir, "insights.md"), []byte("# Insights\n\n- newer insight\n"), 0o644); err != nil {
		t.Fatalf("write newer insights: %v", err)
	}
	if err := os.WriteFile(filepath.Join(currentDir, "insights.md"), []byte(defaultInsightsTemplate), 0o644); err != nil {
		t.Fatalf("write current insights: %v", err)
	}

	insights, err := loadPriorInsights(targetDir, filepath.Join(currentDir, "insights.md"))
	if err != nil {
		t.Fatalf("loadPriorInsights() returned error: %v", err)
	}

	if len(insights) != 2 {
		t.Fatalf("len(loadPriorInsights()) = %d, want 2", len(insights))
	}
	if !strings.Contains(insights[0].Summary, "newer insight") {
		t.Fatalf("first insight summary = %q, want newer insight first", insights[0].Summary)
	}
	if !strings.Contains(insights[1].Summary, "older insight") {
		t.Fatalf("second insight summary = %q, want older insight second", insights[1].Summary)
	}
}

func TestRecordRunStartWritesEvaluatorToResultsLedger(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workdir := filepath.Join(root, "work")
	projectDir := filepath.Join(root, "runtime")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}

	program := `# Program

Objective: record evaluator
Primary evaluator: make manual-lab-up
`
	if err := os.WriteFile(filepath.Join(workdir, defaultProgramFilename), []byte(program), 0o644); err != nil {
		t.Fatalf("write program: %v", err)
	}

	resolver, err := newPromptResolver(workdir, "", projectDir)
	if err != nil {
		t.Fatalf("newPromptResolver() returned error: %v", err)
	}
	artifacts := newAutoresearchArtifacts(workdir, time.Date(2026, time.March, 28, 20, 5, 0, 0, time.UTC))
	resolution, err := resolver.Resolve(artifacts)
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}
	if err := recordRunStart(artifacts, resolution, "run"); err != nil {
		t.Fatalf("recordRunStart() returned error: %v", err)
	}

	data, err := os.ReadFile(artifacts.ResultsLedgerPath)
	if err != nil {
		t.Fatalf("read results ledger: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("results ledger line count = %d, want 1", len(lines))
	}

	var entry resultLedgerEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("Unmarshal(results ledger) returned error: %v", err)
	}
	if entry.Command != "make manual-lab-up" {
		t.Fatalf("ledger command = %q, want make manual-lab-up", entry.Command)
	}
	if entry.Disposition != defaultDispositionPlaceholder {
		t.Fatalf("ledger disposition = %q, want %q", entry.Disposition, defaultDispositionPlaceholder)
	}
}

func TestShouldTriggerCouncilAfterThreshold(t *testing.T) {
	t.Parallel()

	entries := []resultLedgerEntry{
		{Disposition: "keep"},
		{Disposition: "discard"},
		{Disposition: "revert"},
		{Disposition: "blocked"},
	}
	if !shouldTriggerCouncil(entries, 3) {
		t.Fatal("shouldTriggerCouncil() should fire after three consecutive failures")
	}
	if shouldTriggerCouncil(entries, 4) {
		t.Fatal("shouldTriggerCouncil() should not fire before the threshold is met")
	}
}

func TestLoadResultLedgerEntriesSkipsCorruptedLines(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "results.jsonl")
	content := strings.Join([]string{
		`{"timestamp":"2026-03-28T20:19:41Z","disposition":"blocked","command":"go test ./..."}`,
		`not-json-at-all`,
		`{"timestamp":"2026-03-28T20:20:41Z","disposition":"revert","command":"go test ./..."}`,
		"",
	}, "\n")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(results ledger) returned error: %v", err)
	}

	entries, err := loadResultLedgerEntries(path)
	if err != nil {
		t.Fatalf("loadResultLedgerEntries() returned error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("len(loadResultLedgerEntries()) = %d, want 2", len(entries))
	}
	if entries[0].Disposition != "blocked" || entries[1].Disposition != "revert" {
		t.Fatalf("unexpected ledger entries: %+v", entries)
	}
}

func TestShouldTriggerCouncilIgnoresPendingAndPartialEntries(t *testing.T) {
	t.Parallel()

	entries := []resultLedgerEntry{
		{Disposition: "planned"},
		{Disposition: "blocked"},
		{},
		{Disposition: "revert"},
		{Disposition: "blocked"},
	}
	if !shouldTriggerCouncil(entries, 3) {
		t.Fatal("shouldTriggerCouncil() should count consecutive failures across partial/pending entries")
	}
}

func TestShouldTriggerCouncilStopsAtUnknownDisposition(t *testing.T) {
	t.Parallel()

	entries := []resultLedgerEntry{
		{Disposition: "blocked"},
		{Disposition: "custom"},
		{Disposition: "revert"},
		{Disposition: "blocked"},
	}
	if shouldTriggerCouncil(entries, 3) {
		t.Fatal("shouldTriggerCouncil() should stop the streak at unknown dispositions")
	}
}

func TestRunInitCommandScaffoldsWorkspace(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	if err := runInitCommand([]string{"--workdir", workdir}); err != nil {
		t.Fatalf("runInitCommand() returned error: %v", err)
	}

	paths := []string{
		filepath.Join(workdir, defaultProgramFilename),
		filepath.Join(workdir, planningFilename),
		filepath.Join(workdir, targetDirName, runTemplateDirName, "plan.md"),
		filepath.Join(workdir, targetDirName, runTemplateDirName, "execution.md"),
		filepath.Join(workdir, targetDirName, runTemplateDirName, "results.md"),
		filepath.Join(workdir, targetDirName, runTemplateDirName, "insights.md"),
		filepath.Join(workdir, targetDirName, resultsLedgerFilename),
		filepath.Join(workdir, targetDirName, latestContextFilename),
	}
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("scaffold missing %s: %v", path, err)
		}
	}
}
