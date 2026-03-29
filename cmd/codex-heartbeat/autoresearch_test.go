package main

import (
	"bytes"
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

	resolver, err := newPromptResolver(workdir, promptPath, projectDir, false)
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

	resolver, err := newPromptResolver(workdir, "", projectDir, false)
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
	if resolution.Program.Profile != "" || resolution.Program.Model != "" || resolution.Program.ModelReasoningEffort != "" {
		t.Fatalf("Resolve().Program launch settings = %+v, want empty defaults", resolution.Program)
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
	if !strings.Contains(resolution.Text, "agent-paused.lock") {
		t.Fatalf("Resolve().Text missing pause-lock guidance: %q", resolution.Text)
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

	resolver, err := newPromptResolver(workdir, "", projectDir, false)
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

	resolver, err := newPromptResolver(workdir, "", projectDir, false)
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

func TestPromptResolverPlanningMode(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workdir := filepath.Join(root, "work")
	projectDir := filepath.Join(root, "runtime")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}

	program := `# Program

Objective: build the plan before coding
Primary evaluator: rg -n "TODO|FIXME" .
Prompt mode: planning

## Notes

- Deepen the implementation plan before making broad code changes.
`
	if err := os.WriteFile(filepath.Join(workdir, defaultProgramFilename), []byte(program), 0o644); err != nil {
		t.Fatalf("write program: %v", err)
	}

	resolver, err := newPromptResolver(workdir, "", projectDir, false)
	if err != nil {
		t.Fatalf("newPromptResolver() returned error: %v", err)
	}

	artifacts := newAutoresearchArtifacts(workdir, time.Date(2026, time.March, 29, 1, 0, 0, 0, time.UTC))
	resolution, err := resolver.Resolve(artifacts)
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	if !strings.Contains(resolution.Text, "refine the goal, deepen the plan") {
		t.Fatalf("Resolve().Text missing planning-mode guidance: %q", resolution.Text)
	}
	if !strings.Contains(resolution.Text, "Planning history path") {
		t.Fatalf("Resolve().Text missing planning history path: %q", resolution.Text)
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

	resolver, err := newPromptResolver(workdir, "", projectDir, false)
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

func TestLoadProgramConfigParsesLaunchOverrides(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, defaultProgramFilename)
	program := `# Program

Objective: use program launch settings
Primary evaluator: go test ./cmd/codex-heartbeat
Profile: safe-research
Model: gpt-5.3-codex-spark
Model reasoning effort: high
`
	if err := os.WriteFile(path, []byte(program), 0o644); err != nil {
		t.Fatalf("write program: %v", err)
	}

	cfg, err := loadProgramConfig(path)
	if err != nil {
		t.Fatalf("loadProgramConfig() returned error: %v", err)
	}
	if cfg.Profile != "safe-research" {
		t.Fatalf("Profile = %q, want safe-research", cfg.Profile)
	}
	if cfg.Model != "gpt-5.3-codex-spark" {
		t.Fatalf("Model = %q, want gpt-5.3-codex-spark", cfg.Model)
	}
	if cfg.ModelReasoningEffort != "high" {
		t.Fatalf("ModelReasoningEffort = %q, want high", cfg.ModelReasoningEffort)
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
	if err := os.WriteFile(filepath.Join(currentDir, "insights.md"), []byte(embeddedTemplate("insights.md")), 0o644); err != nil {
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
Profile: safe-research
Model: gpt-5.3-codex-spark
Model reasoning effort: high
`
	if err := os.WriteFile(filepath.Join(workdir, defaultProgramFilename), []byte(program), 0o644); err != nil {
		t.Fatalf("write program: %v", err)
	}

	resolver, err := newPromptResolver(workdir, "", projectDir, true)
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
	if !strings.Contains(entry.Notes, "council_policy=`frequent`") {
		t.Fatalf("ledger notes = %q, want frequent council policy", entry.Notes)
	}
	if !strings.Contains(entry.Notes, "launch_settings=`profile=safe-research, model=gpt-5.3-codex-spark, model_reasoning_effort=high`") {
		t.Fatalf("ledger notes = %q, want recorded launch settings", entry.Notes)
	}
}

func TestPromptResolverWritesLaunchSettingsToLatestContext(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workdir := filepath.Join(root, "work")
	projectDir := filepath.Join(root, "runtime")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}

	program := `# Program

Objective: record launch settings in context
Primary evaluator: go test ./cmd/codex-heartbeat
Profile: safe-research
Model: gpt-5.3-codex-spark
Model reasoning effort: high
`
	if err := os.WriteFile(filepath.Join(workdir, defaultProgramFilename), []byte(program), 0o644); err != nil {
		t.Fatalf("write program: %v", err)
	}

	resolver, err := newPromptResolver(workdir, "", projectDir, false)
	if err != nil {
		t.Fatalf("newPromptResolver() returned error: %v", err)
	}
	artifacts := newAutoresearchArtifacts(workdir, time.Date(2026, time.March, 28, 20, 7, 0, 0, time.UTC))
	if _, err := resolver.Resolve(artifacts); err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	data, err := os.ReadFile(artifacts.LatestContextPath)
	if err != nil {
		t.Fatalf("read latest context: %v", err)
	}
	if !strings.Contains(string(data), "- Launch settings: `profile=safe-research, model=gpt-5.3-codex-spark, model_reasoning_effort=high`") {
		t.Fatalf("latest context missing launch settings: %q", string(data))
	}
}

func TestPromptResolverCouncilFlagUsesFrequentCouncilPolicy(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	workdir := filepath.Join(root, "work")
	projectDir := filepath.Join(root, "runtime")
	if err := os.MkdirAll(workdir, 0o755); err != nil {
		t.Fatalf("mkdir workdir: %v", err)
	}

	program := `# Program

Objective: fix the flaky evaluator
Primary evaluator: go test ./...
`
	if err := os.WriteFile(filepath.Join(workdir, defaultProgramFilename), []byte(program), 0o644); err != nil {
		t.Fatalf("write program: %v", err)
	}

	resolver, err := newPromptResolver(workdir, "", projectDir, true)
	if err != nil {
		t.Fatalf("newPromptResolver() returned error: %v", err)
	}

	artifacts := newAutoresearchArtifacts(workdir, time.Date(2026, time.March, 28, 20, 6, 0, 0, time.UTC))
	resolution, err := resolver.Resolve(artifacts)
	if err != nil {
		t.Fatalf("Resolve() returned error: %v", err)
	}

	if !resolution.CouncilRequested {
		t.Fatal("Resolve() should record council mode when --council is set")
	}
	if !resolution.CouncilTriggered {
		t.Fatal("Resolve() should trigger council guidance when --council is set")
	}
	if !strings.Contains(resolution.Text, "Use the 3-agent council at many steps") {
		t.Fatalf("Resolve().Text missing frequent council policy: %q", resolution.Text)
	}
	if !strings.Contains(resolution.Text, "gpt-5.4") || !strings.Contains(resolution.Text, "gpt-5.3-codex-spark") {
		t.Fatalf("Resolve().Text missing requested council model split: %q", resolution.Text)
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

func TestEnsureAutoresearchWorkspaceScaffoldsWorkspace(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	if warning, err := ensureAutoresearchWorkspace(workdir); err != nil {
		t.Fatalf("ensureAutoresearchWorkspace() returned error: %v", err)
	} else if warning != "" {
		t.Fatalf("ensureAutoresearchWorkspace() warning = %q, want empty warning for a blank workspace", warning)
	}

	paths := []string{
		filepath.Join(workdir, defaultProgramFilename),
		filepath.Join(workdir, planningFilename),
		filepath.Join(workdir, targetDirName, planningHistoryFilename),
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

func TestEnsureAutoresearchWorkspaceSeedsPlanningHistory(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	if _, err := ensureAutoresearchWorkspace(workdir); err != nil {
		t.Fatalf("ensureAutoresearchWorkspace() returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workdir, targetDirName, planningHistoryFilename))
	if err != nil {
		t.Fatalf("read planning history scaffold: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# Planning History") {
		t.Fatalf("planning history scaffold missing title: %q", content)
	}
	if !strings.Contains(content, "durable planning memory") {
		t.Fatalf("planning history scaffold missing durable-memory guidance: %q", content)
	}
}

func TestEnsureAutoresearchWorkspaceSeedsPlanningTaskList(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	if _, err := ensureAutoresearchWorkspace(workdir); err != nil {
		t.Fatalf("ensureAutoresearchWorkspace() returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workdir, planningFilename))
	if err != nil {
		t.Fatalf("read planning scaffold: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "## Task List") {
		t.Fatalf("planning scaffold missing task list section: %q", content)
	}
	if !strings.Contains(content, "- [ ] Pick one bounded hypothesis for the next cycle.") {
		t.Fatalf("planning scaffold missing bounded-hypothesis checkbox: %q", content)
	}
	if !strings.Contains(content, "- [ ] Name one primary evaluator before changing code.") {
		t.Fatalf("planning scaffold missing evaluator checkbox: %q", content)
	}
}

func TestEnsureAutoresearchWorkspaceSeedsPlanningGuardrails(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	if _, err := ensureAutoresearchWorkspace(workdir); err != nil {
		t.Fatalf("ensureAutoresearchWorkspace() returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(workdir, planningFilename))
	if err != nil {
		t.Fatalf("read planning scaffold: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "## Blocked / Non-Goals") {
		t.Fatalf("planning scaffold missing blocked/non-goals section: %q", content)
	}
	if !strings.Contains(content, "## Acceptance Criteria") {
		t.Fatalf("planning scaffold missing acceptance criteria section: %q", content)
	}
	if !strings.Contains(content, "- [ ] Call out any missing capability that keeps the current parity or success claim false.") {
		t.Fatalf("planning scaffold missing parity-guardrail checkbox: %q", content)
	}
	if !strings.Contains(content, "- [ ] Define the observable result that would make this cycle a keep.") {
		t.Fatalf("planning scaffold missing acceptance checkbox: %q", content)
	}
}

func TestEnsureAutoresearchWorkspaceWarnsOnPartialScaffoldWithoutOverwriting(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	programPath := filepath.Join(workdir, defaultProgramFilename)
	existingProgram := "# Program\n\nObjective: preserve me.\n"
	if err := os.WriteFile(programPath, []byte(existingProgram), 0o644); err != nil {
		t.Fatalf("write program: %v", err)
	}

	warning, err := ensureAutoresearchWorkspace(workdir)
	if err != nil {
		t.Fatalf("ensureAutoresearchWorkspace() returned error: %v", err)
	}
	if !strings.Contains(warning, "partially present") {
		t.Fatalf("ensureAutoresearchWorkspace() warning = %q, want partial scaffold warning", warning)
	}
	if !strings.Contains(warning, "PLANNING.md") {
		t.Fatalf("ensureAutoresearchWorkspace() warning = %q, want missing scaffold path list", warning)
	}

	data, err := os.ReadFile(programPath)
	if err != nil {
		t.Fatalf("read program: %v", err)
	}
	if string(data) != existingProgram {
		t.Fatalf("program.md was overwritten: got %q want %q", string(data), existingProgram)
	}
	if _, err := os.Stat(filepath.Join(workdir, planningFilename)); err != nil {
		t.Fatalf("PLANNING.md missing after ensureAutoresearchWorkspace(): %v", err)
	}
	if _, err := os.Stat(filepath.Join(workdir, targetDirName, planningHistoryFilename)); err != nil {
		t.Fatalf("target/PLANNING_HISTORY.md missing after ensureAutoresearchWorkspace(): %v", err)
	}
}

func TestEnsureAutoresearchWorkspaceWithSurveyUsesAnswers(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	input := strings.NewReader("Ship a planning-mode refactor\nmake test\nMap the architecture and dead ends first\n1\n")
	var output bytes.Buffer

	if warning, err := ensureAutoresearchWorkspaceWithSurvey(workdir, input, &output, true); err != nil {
		t.Fatalf("ensureAutoresearchWorkspaceWithSurvey() returned error: %v", err)
	} else if warning != "" {
		t.Fatalf("ensureAutoresearchWorkspaceWithSurvey() warning = %q, want empty warning", warning)
	}

	programData, err := os.ReadFile(filepath.Join(workdir, defaultProgramFilename))
	if err != nil {
		t.Fatalf("read program scaffold: %v", err)
	}
	if !strings.Contains(string(programData), "Objective: Ship a planning-mode refactor") {
		t.Fatalf("program scaffold missing surveyed objective: %q", string(programData))
	}
	if !strings.Contains(string(programData), "Prompt mode: planning") {
		t.Fatalf("program scaffold missing planning mode: %q", string(programData))
	}

	planningData, err := os.ReadFile(filepath.Join(workdir, planningFilename))
	if err != nil {
		t.Fatalf("read planning scaffold: %v", err)
	}
	if !strings.Contains(string(planningData), "- [ ] Map the architecture and dead ends first") {
		t.Fatalf("planning scaffold missing surveyed deep dive: %q", string(planningData))
	}
}

func TestArchiveCompletedPlanningTasksMovesCheckedItems(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	artifacts := newAutoresearchArtifacts(workdir, time.Date(2026, time.March, 29, 1, 5, 0, 0, time.UTC))
	if err := os.MkdirAll(filepath.Dir(artifacts.PlanningHistoryPath), 0o755); err != nil {
		t.Fatalf("mkdir planning history dir: %v", err)
	}
	planning := `# Planning

## Task List

- [x] Investigated the legacy retry path.
- [ ] Implement the retry fix.

## Open Questions

- [x] Decide whether the stale cache path still matters.
`
	if err := os.WriteFile(filepath.Join(workdir, planningFilename), []byte(planning), 0o644); err != nil {
		t.Fatalf("write planning: %v", err)
	}

	if err := archiveCompletedPlanningTasks(artifacts, time.Date(2026, time.March, 29, 1, 6, 0, 0, time.UTC)); err != nil {
		t.Fatalf("archiveCompletedPlanningTasks() returned error: %v", err)
	}

	updatedPlanning, err := os.ReadFile(filepath.Join(workdir, planningFilename))
	if err != nil {
		t.Fatalf("read planning: %v", err)
	}
	if strings.Contains(string(updatedPlanning), "- [x] Investigated the legacy retry path.") {
		t.Fatalf("updated planning still contains archived task: %q", string(updatedPlanning))
	}
	if !strings.Contains(string(updatedPlanning), "- [ ] Implement the retry fix.") {
		t.Fatalf("updated planning lost active task: %q", string(updatedPlanning))
	}

	history, err := os.ReadFile(artifacts.PlanningHistoryPath)
	if err != nil {
		t.Fatalf("read planning history: %v", err)
	}
	if !strings.Contains(string(history), "Investigated the legacy retry path") {
		t.Fatalf("planning history missing archived task: %q", string(history))
	}
	if !strings.Contains(string(history), "### Task List") {
		t.Fatalf("planning history missing section heading: %q", string(history))
	}
}

func TestHeartbeatPauseStateHonorsExistingPauseLock(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	artifacts := newAutoresearchArtifacts(workdir, time.Date(2026, time.March, 29, 2, 0, 0, 0, time.UTC))
	if err := os.WriteFile(artifacts.PauseLockPath, []byte("paused\n"), 0o644); err != nil {
		t.Fatalf("write pause lock: %v", err)
	}

	paused, reason, err := heartbeatPauseState(artifacts)
	if err != nil {
		t.Fatalf("heartbeatPauseState() returned error: %v", err)
	}
	if !paused {
		t.Fatal("heartbeatPauseState() should pause when the lock file exists")
	}
	if reason != "pause_lock_present" {
		t.Fatalf("heartbeatPauseState() reason = %q, want pause_lock_present", reason)
	}
}

func TestHeartbeatPauseStateCreatesPauseLockFromCompleteDisposition(t *testing.T) {
	t.Parallel()

	workdir := t.TempDir()
	artifacts := newAutoresearchArtifacts(workdir, time.Date(2026, time.March, 29, 2, 1, 0, 0, time.UTC))
	entry := resultLedgerEntry{
		Timestamp:   "2026-03-29T02:01:00Z",
		Command:     "go test ./...",
		Outcome:     "pass",
		Disposition: "complete",
	}
	if err := appendResultLedgerEntry(artifacts.ResultsLedgerPath, entry); err != nil {
		t.Fatalf("appendResultLedgerEntry() returned error: %v", err)
	}

	paused, reason, err := heartbeatPauseState(artifacts)
	if err != nil {
		t.Fatalf("heartbeatPauseState() returned error: %v", err)
	}
	if !paused {
		t.Fatal("heartbeatPauseState() should pause when the latest disposition is complete")
	}
	if reason != "objective_complete" {
		t.Fatalf("heartbeatPauseState() reason = %q, want objective_complete", reason)
	}

	lock, err := os.ReadFile(artifacts.PauseLockPath)
	if err != nil {
		t.Fatalf("read pause lock: %v", err)
	}
	if !strings.Contains(string(lock), "objective achieved") {
		t.Fatalf("pause lock missing objective-achieved note: %q", string(lock))
	}
}
