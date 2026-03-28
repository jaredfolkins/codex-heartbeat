package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	defaultProgramFilename        = "program.md"
	planningFilename              = "PLANNING.md"
	targetDirName                 = "target"
	runDirPrefix                  = "run-"
	runTemplateDirName            = "templates"
	resultsLedgerFilename         = "results.jsonl"
	latestContextFilename         = "latest-context.md"
	defaultCouncilAfterFailures   = 3
	maxProgramBytes               = 64 * 1024
	maxInsightFiles               = 12
	maxInsightBytesPerFile        = 4 * 1024
	maxLatestContextBytes         = 12 * 1024
	defaultHypothesisPlaceholder  = "Fill in one falsifiable hypothesis before you change code."
	defaultDispositionPlaceholder = "planned"
)

const (
	promptSourceCLI      = "cli_prompt"
	promptSourceProgram  = "program_md"
	promptSourceEmbedded = "embedded"
)

type promptMode string

const (
	promptModeAutoresearch    promptMode = "autoresearch"
	promptModeManualTestFirst promptMode = "manual-test-first"
)

type promptResolver struct {
	workdir  string
	explicit promptSource
	council  bool
}

type promptResolution struct {
	Text             string
	Source           string
	SourcePath       string
	UsedCache        bool
	Program          programConfig
	CouncilRequested bool
	CouncilTriggered bool
}

type programConfig struct {
	Path                 string
	Objective            string
	PrimaryEvaluator     string
	PromptMode           promptMode
	CouncilAfterFailures int
	CheckpointCommits    bool
	Body                 string
}

type autoresearchArtifacts struct {
	TargetDir         string
	TemplateDir       string
	RunID             string
	RunDir            string
	PlanPath          string
	ExecutionPath     string
	ResultsPath       string
	InsightsPath      string
	LatestContextPath string
	ResultsLedgerPath string
}

type resultLedgerEntry struct {
	Timestamp   string `json:"timestamp"`
	RunID       string `json:"run_id,omitempty"`
	Hypothesis  string `json:"hypothesis,omitempty"`
	Command     string `json:"command,omitempty"`
	Outcome     string `json:"outcome,omitempty"`
	Disposition string `json:"disposition,omitempty"`
	Notes       string `json:"notes,omitempty"`
}

type priorInsight struct {
	Path    string
	Summary string
}

func newPromptResolver(workdir, promptPath, projectDir string, council bool) (promptResolver, error) {
	explicit, err := newPromptSource(promptPath, projectDir)
	if err != nil {
		return promptResolver{}, err
	}
	return promptResolver{
		workdir:  workdir,
		explicit: explicit,
		council:  council,
	}, nil
}

func (r promptResolver) Validate() error {
	_, err := r.resolveBasePrompt()
	return err
}

func (r promptResolver) Resolve(artifacts autoresearchArtifacts) (promptResolution, error) {
	base, err := r.resolveBasePrompt()
	if err != nil {
		return promptResolution{}, err
	}

	if artifacts.TargetDir == "" {
		return promptResolution{
			Text:             base.text,
			Source:           base.source,
			SourcePath:       base.sourcePath,
			UsedCache:        base.usedCache,
			Program:          base.program,
			CouncilRequested: r.council,
		}, nil
	}

	entries, err := loadResultLedgerEntries(artifacts.ResultsLedgerPath)
	if err != nil {
		return promptResolution{}, err
	}

	threshold := base.program.CouncilAfterFailures
	if threshold <= 0 {
		threshold = defaultCouncilAfterFailures
	}
	councilTriggered := r.council || shouldTriggerCouncil(entries, threshold)

	latestContext, err := buildLatestContext(artifacts, base.program, r.council, entries)
	if err != nil {
		return promptResolution{}, err
	}
	if err := writeLatestContext(artifacts, latestContext); err != nil {
		return promptResolution{}, err
	}

	text := base.text
	if base.source != promptSourceCLI {
		text = renderPromptTemplate(defaultPrompt, buildPromptTemplateVars(base.program, artifacts, latestContext, councilTriggered, r.council))
	} else if r.council {
		text = strings.TrimSpace("Council mode:\n" + buildCouncilInstruction(threshold, councilTriggered, r.council) + "\n\n" + base.text)
	}

	resolution := promptResolution{
		Text:             text,
		Source:           base.source,
		SourcePath:       base.sourcePath,
		UsedCache:        base.usedCache,
		Program:          base.program,
		CouncilRequested: r.council,
		CouncilTriggered: councilTriggered,
	}

	if err := ensureRunArtifacts(artifacts, resolution); err != nil {
		return promptResolution{}, err
	}

	return resolution, nil
}

type basePromptResolution struct {
	text       string
	source     string
	sourcePath string
	usedCache  bool
	program    programConfig
}

func (r promptResolver) resolveBasePrompt() (basePromptResolution, error) {
	if strings.TrimSpace(r.explicit.path) != "" {
		text, usedCache, err := r.explicit.ResolveWithMetadata()
		if err != nil {
			return basePromptResolution{}, err
		}
		return basePromptResolution{
			text:       text,
			source:     promptSourceCLI,
			sourcePath: r.explicit.path,
			usedCache:  usedCache,
			program:    defaultProgramConfig(""),
		}, nil
	}

	programPath := filepath.Join(r.workdir, defaultProgramFilename)
	program, err := loadProgramConfig(programPath)
	if err == nil {
		return basePromptResolution{
			source:     promptSourceProgram,
			sourcePath: programPath,
			program:    program,
		}, nil
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) && !errors.Is(err, errProgramTooLarge) && !errors.Is(err, errEmptyProgram) {
		return basePromptResolution{}, err
	}

	program = defaultProgramConfig(programPath)
	return basePromptResolution{
		source:     promptSourceEmbedded,
		sourcePath: "embedded heartbeat.md",
		program:    program,
	}, nil
}

var (
	errProgramTooLarge = errors.New("program.md exceeds the size cap")
	errEmptyProgram    = errors.New("program.md is empty")
)

func defaultProgramConfig(path string) programConfig {
	return programConfig{
		Path:                 path,
		Objective:            "Make one measurable improvement in the current workspace.",
		PrimaryEvaluator:     "Choose one concrete evaluator command or manual validation step before you edit, then reuse it for the cycle.",
		PromptMode:           promptModeAutoresearch,
		CouncilAfterFailures: defaultCouncilAfterFailures,
		CheckpointCommits:    false,
		Body: strings.TrimSpace(`
## Program

No human-authored program was found in this workspace.

- Infer the smallest useful objective from the repo state.
- Reuse one evaluator for the full cycle.
- Update the run artifacts under target/ so the next heartbeat has memory.
`),
	}
}

func loadProgramConfig(path string) (programConfig, error) {
	info, err := os.Stat(path)
	if err != nil {
		return programConfig{}, err
	}
	if info.IsDir() {
		return programConfig{}, fmt.Errorf("%s is a directory", path)
	}
	if info.Size() > maxProgramBytes {
		return programConfig{}, errProgramTooLarge
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return programConfig{}, err
	}

	text := strings.TrimSpace(string(data))
	if text == "" {
		return programConfig{}, errEmptyProgram
	}

	cfg := defaultProgramConfig(path)
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	bodyStart := 0
	foundMetadata := false

	for idx, raw := range lines {
		line := strings.TrimSpace(raw)
		if idx == 0 && strings.HasPrefix(line, "#") {
			bodyStart = idx + 1
			continue
		}
		if line == "" {
			if foundMetadata {
				bodyStart = idx + 1
				break
			}
			bodyStart = idx + 1
			continue
		}
		if strings.HasPrefix(line, "##") {
			bodyStart = idx
			break
		}
		key, value, ok := parseProgramMetadataLine(line)
		if !ok {
			bodyStart = idx
			break
		}
		foundMetadata = true
		applyProgramMetadata(&cfg, key, value)
		bodyStart = idx + 1
	}

	body := strings.TrimSpace(strings.Join(lines[bodyStart:], "\n"))
	if body != "" {
		cfg.Body = body
	}
	return cfg, nil
}

func parseProgramMetadataLine(line string) (string, string, bool) {
	before, after, ok := strings.Cut(line, ":")
	if !ok {
		return "", "", false
	}
	key := strings.TrimSpace(before)
	value := strings.TrimSpace(after)
	if key == "" || value == "" {
		return "", "", false
	}
	return normalizeProgramMetadataKey(key), value, true
}

func normalizeProgramMetadataKey(key string) string {
	key = strings.ToLower(strings.TrimSpace(key))
	replacer := strings.NewReplacer(" ", "", "-", "", "_", "")
	return replacer.Replace(key)
}

func applyProgramMetadata(cfg *programConfig, key, value string) {
	switch key {
	case "objective":
		cfg.Objective = value
	case "primaryevaluator", "evaluator", "evaluatorcommand":
		cfg.PrimaryEvaluator = value
	case "promptmode", "mode":
		switch strings.ToLower(strings.TrimSpace(value)) {
		case string(promptModeManualTestFirst):
			cfg.PromptMode = promptModeManualTestFirst
		default:
			cfg.PromptMode = promptModeAutoresearch
		}
	case "councilafter", "councilafterfailures", "councilthreshold":
		if parsed, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && parsed > 0 {
			cfg.CouncilAfterFailures = parsed
		}
	case "checkpointcommits", "savepointcommits", "gitcheckpoints":
		if parsed, err := strconv.ParseBool(strings.TrimSpace(value)); err == nil {
			cfg.CheckpointCommits = parsed
		}
	}
}

func newAutoresearchArtifacts(workdir string, startedAt time.Time) autoresearchArtifacts {
	targetDir := filepath.Join(workdir, targetDirName)
	runID := runDirPrefix + startedAt.UTC().Format("20060102T150405Z")
	runDir := filepath.Join(targetDir, runID)

	return autoresearchArtifacts{
		TargetDir:         targetDir,
		TemplateDir:       filepath.Join(targetDir, runTemplateDirName),
		RunID:             runID,
		RunDir:            runDir,
		PlanPath:          filepath.Join(runDir, "plan.md"),
		ExecutionPath:     filepath.Join(runDir, "execution.md"),
		ResultsPath:       filepath.Join(runDir, "results.md"),
		InsightsPath:      filepath.Join(runDir, "insights.md"),
		LatestContextPath: filepath.Join(targetDir, latestContextFilename),
		ResultsLedgerPath: filepath.Join(targetDir, resultsLedgerFilename),
	}
}

func buildPromptTemplateVars(program programConfig, artifacts autoresearchArtifacts, latestContext string, councilTriggered bool, councilRequested bool) map[string]string {
	threshold := program.CouncilAfterFailures
	if threshold <= 0 {
		threshold = defaultCouncilAfterFailures
	}

	modeInstruction := "Stop only after you have run the evaluator, recorded the result, and decided keep, discard, or revert."
	if program.PromptMode == promptModeManualTestFirst {
		modeInstruction = "Prepare the next candidate fix and the exact validation steps, then stop before the final human gate. Do not take the final apply/ship step."
	}

	checkpointInstruction := "Do not create a save-point commit unless the human explicitly asks."
	if program.CheckpointCommits {
		checkpointInstruction = "After meaningful progress, create an intentional save-point commit that explains what changed and why."
	}

	councilInstruction := buildCouncilInstruction(threshold, councilTriggered, councilRequested)

	return map[string]string{
		"OBJECTIVE":               strings.TrimSpace(program.Objective),
		"PRIMARY_EVALUATOR":       strings.TrimSpace(program.PrimaryEvaluator),
		"PROMPT_MODE":             string(program.PromptMode),
		"COUNCIL_AFTER_FAILURES":  strconv.Itoa(threshold),
		"COUNCIL_INSTRUCTION":     councilInstruction,
		"CHECKPOINT_INSTRUCTION":  checkpointInstruction,
		"MANUAL_GATE_INSTRUCTION": modeInstruction,
		"LATEST_CONTEXT_PATH":     artifacts.LatestContextPath,
		"RUN_DIR":                 artifacts.RunDir,
		"PROGRAM_BODY":            strings.TrimSpace(program.Body),
		"LATEST_CONTEXT":          strings.TrimSpace(latestContext),
	}
}

func buildCouncilInstruction(threshold int, councilTriggered bool, councilRequested bool) string {
	if councilRequested {
		return "Use the 3-agent council at many steps in the autoresearch loop: baseline framing, next-hypothesis selection, and post-evaluator interpretation. Keep the root agent on `gpt-5.4` with `xhigh` reasoning. Use `gpt-5.3-codex-spark` with `high` reasoning for the three sub-agents. Still keep each cycle bounded to one hypothesis and one primary evaluator."
	}
	if councilTriggered {
		return fmt.Sprintf("Recent results indicate a stalled loop. Before choosing the next hypothesis, use the 3-agent council because the failure streak reached the threshold of %d.", threshold)
	}
	return fmt.Sprintf("Do not start with the 3-agent council. Use it only if you are blocked or the recent failure streak reaches %d.", threshold)
}

func renderPromptTemplate(template string, vars map[string]string) string {
	replacements := []string{}
	for key, value := range vars {
		replacements = append(replacements, "{{"+key+"}}", strings.TrimSpace(value))
	}
	replacer := strings.NewReplacer(replacements...)
	return strings.TrimSpace(replacer.Replace(template))
}

func ensureRunArtifacts(artifacts autoresearchArtifacts, resolution promptResolution) error {
	if err := os.MkdirAll(artifacts.RunDir, 0o755); err != nil {
		return fmt.Errorf("create run dir: %w", err)
	}

	vars := buildArtifactTemplateVars(artifacts, resolution)
	files := []struct {
		path         string
		templatePath string
		fallback     string
	}{
		{path: artifacts.PlanPath, templatePath: filepath.Join(artifacts.TemplateDir, "plan.md"), fallback: defaultPlanTemplate},
		{path: artifacts.ExecutionPath, templatePath: filepath.Join(artifacts.TemplateDir, "execution.md"), fallback: defaultExecutionTemplate},
		{path: artifacts.ResultsPath, templatePath: filepath.Join(artifacts.TemplateDir, "results.md"), fallback: defaultResultsTemplate},
		{path: artifacts.InsightsPath, templatePath: filepath.Join(artifacts.TemplateDir, "insights.md"), fallback: defaultInsightsTemplate},
	}

	for _, file := range files {
		content := renderPromptTemplate(loadTemplateOrDefault(file.templatePath, file.fallback), vars) + "\n"
		if err := writeFileIfMissing(file.path, []byte(content), 0o644); err != nil {
			return err
		}
	}

	return nil
}

func buildArtifactTemplateVars(artifacts autoresearchArtifacts, resolution promptResolution) map[string]string {
	program := resolution.Program
	threshold := program.CouncilAfterFailures
	if threshold <= 0 {
		threshold = defaultCouncilAfterFailures
	}
	return map[string]string{
		"RUN_ID":                 artifacts.RunID,
		"RUN_DIR":                artifacts.RunDir,
		"OBJECTIVE":              program.Objective,
		"PRIMARY_EVALUATOR":      program.PrimaryEvaluator,
		"PROMPT_SOURCE":          resolution.Source,
		"PROMPT_SOURCE_PATH":     resolution.SourcePath,
		"PROMPT_MODE":            string(program.PromptMode),
		"COUNCIL_POLICY":         councilPolicyLabel(resolution.CouncilRequested),
		"COUNCIL_AFTER_FAILURES": strconv.Itoa(threshold),
		"COUNCIL_TRIGGERED":      strconv.FormatBool(resolution.CouncilTriggered),
		"CHECKPOINT_COMMITS":     strconv.FormatBool(program.CheckpointCommits),
		"LATEST_CONTEXT_PATH":    artifacts.LatestContextPath,
		"HYPOTHESIS_PLACEHOLDER": defaultHypothesisPlaceholder,
		"RESULTS_LEDGER_PATH":    artifacts.ResultsLedgerPath,
	}
}

func loadTemplateOrDefault(path, fallback string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func writeFileIfMissing(path string, data []byte, perm fs.FileMode) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return writeFileAtomic(path, data, perm)
}

func writeFileAtomic(path string, data []byte, perm fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	current, err := os.ReadFile(path)
	if err == nil && bytes.Equal(current, data) {
		return nil
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".codex-heartbeat-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, path)
}

func writeLatestContext(artifacts autoresearchArtifacts, latestContext string) error {
	return writeFileAtomic(artifacts.LatestContextPath, []byte(strings.TrimSpace(latestContext)+"\n"), 0o644)
}

func buildLatestContext(artifacts autoresearchArtifacts, program programConfig, councilRequested bool, entries []resultLedgerEntry) (string, error) {
	insights, err := loadPriorInsights(artifacts.TargetDir, artifacts.InsightsPath)
	if err != nil {
		return "", err
	}

	threshold := program.CouncilAfterFailures
	if threshold <= 0 {
		threshold = defaultCouncilAfterFailures
	}
	streak := consecutiveFailureCount(entries)

	lines := []string{
		"# Latest Context",
		"",
		fmt.Sprintf("- Objective: %s", strings.TrimSpace(program.Objective)),
		fmt.Sprintf("- Primary evaluator: `%s`", strings.TrimSpace(program.PrimaryEvaluator)),
		fmt.Sprintf("- Prompt mode: `%s`", program.PromptMode),
		fmt.Sprintf("- Council policy: `%s`", councilPolicyLabel(councilRequested)),
		fmt.Sprintf("- Recent failure streak: %d / %d", streak, threshold),
		"",
		"## Recent Ledger",
	}

	recentLedger := recentLedgerSummary(entries, 5)
	if len(recentLedger) == 0 {
		lines = append(lines, "- No prior result entries recorded yet.")
	} else {
		lines = append(lines, recentLedger...)
	}

	lines = append(lines, "", "## Prior Insights")
	if len(insights) == 0 {
		lines = append(lines, "- No prior `target/*/insights.md` artifacts were found.")
	} else {
		for _, insight := range insights {
			lines = append(lines, fmt.Sprintf("- %s: %s", shortenInsightPath(artifacts.TargetDir, insight.Path), insight.Summary))
		}
	}

	context := strings.TrimSpace(strings.Join(lines, "\n"))
	if len(context) > maxLatestContextBytes {
		context = context[:maxLatestContextBytes]
		context = strings.TrimRight(context, "\n")
		context += "\n\n- Context truncated to stay within the prompt budget."
	}
	return context, nil
}

func loadPriorInsights(targetDir, currentInsightsPath string) ([]priorInsight, error) {
	pattern := filepath.Join(targetDir, runDirPrefix+"*", "insights.md")
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(paths)))

	insights := make([]priorInsight, 0, min(len(paths), maxInsightFiles))
	for _, path := range paths {
		if samePath(path, currentInsightsPath) {
			continue
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if len(data) > maxInsightBytesPerFile {
			data = data[:maxInsightBytesPerFile]
		}

		summary := summarizeInsightContent(string(data))
		if summary == "" {
			continue
		}

		insights = append(insights, priorInsight{
			Path:    path,
			Summary: summary,
		})
		if len(insights) >= maxInsightFiles {
			break
		}
	}
	return insights, nil
}

func summarizeInsightContent(content string) string {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	parts := []string{}
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			parts = append(parts, strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* ")))
		} else if len(parts) < 3 {
			parts = append(parts, line)
		}
		if len(parts) >= 4 {
			break
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func shortenInsightPath(targetDir, insightPath string) string {
	rel, err := filepath.Rel(targetDir, insightPath)
	if err != nil {
		return insightPath
	}
	return rel
}

func loadResultLedgerEntries(path string) ([]resultLedgerEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	entries := []resultLedgerEntry{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry resultLedgerEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func appendResultLedgerEntry(path string, entry resultLedgerEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

func recordRunStart(artifacts autoresearchArtifacts, resolution promptResolution, commandName string) error {
	note := fmt.Sprintf("started via `%s`; prompt source=`%s`; mode=`%s`; council_policy=`%s`; council_triggered=%t", commandName, resolution.Source, resolution.Program.PromptMode, councilPolicyLabel(resolution.CouncilRequested), resolution.CouncilTriggered)
	if resolution.UsedCache {
		note += "; prompt cache fallback used"
	}
	if err := appendExecutionNote(artifacts.ExecutionPath, note); err != nil {
		return err
	}

	return appendResultLedgerEntry(artifacts.ResultsLedgerPath, resultLedgerEntry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		RunID:       artifacts.RunID,
		Hypothesis:  defaultHypothesisPlaceholder,
		Command:     resolution.Program.PrimaryEvaluator,
		Outcome:     "pending",
		Disposition: defaultDispositionPlaceholder,
		Notes:       note,
	})
}

func councilPolicyLabel(councilRequested bool) string {
	if councilRequested {
		return "frequent"
	}
	return "fallback"
}

func appendExecutionNote(path, note string) error {
	if strings.TrimSpace(note) == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	line := fmt.Sprintf("- %s %s\n", time.Now().UTC().Format(time.RFC3339), strings.TrimSpace(note))
	_, err = file.WriteString(line)
	return err
}

func recentLedgerSummary(entries []resultLedgerEntry, limit int) []string {
	if limit <= 0 || len(entries) == 0 {
		return nil
	}

	summary := []string{}
	for idx := len(entries) - 1; idx >= 0 && len(summary) < limit; idx-- {
		entry := entries[idx]
		disposition := strings.TrimSpace(entry.Disposition)
		if disposition == "" {
			disposition = "unknown"
		}
		command := strings.TrimSpace(entry.Command)
		if command == "" {
			command = "n/a"
		}
		outcome := strings.TrimSpace(entry.Outcome)
		if outcome == "" {
			outcome = "n/a"
		}
		note := strings.TrimSpace(entry.Notes)
		if note != "" {
			note = " | " + note
		}
		summary = append(summary, fmt.Sprintf("- `%s` via `%s`: %s%s", disposition, command, outcome, note))
	}
	return summary
}

func shouldTriggerCouncil(entries []resultLedgerEntry, threshold int) bool {
	if threshold <= 0 {
		return false
	}
	return consecutiveFailureCount(entries) >= threshold
}

func consecutiveFailureCount(entries []resultLedgerEntry) int {
	count := 0
	for idx := len(entries) - 1; idx >= 0; idx-- {
		disposition := normalizeLedgerValue(entries[idx].Disposition)
		if disposition == "" || disposition == "planned" || disposition == "pending" || disposition == "prompted" {
			continue
		}
		if isFailureDisposition(disposition) {
			count++
			continue
		}
		if disposition == "keep" || disposition == "success" {
			break
		}
		break
	}
	return count
}

func normalizeLedgerValue(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	replacer := strings.NewReplacer(" ", "_", "-", "_")
	return replacer.Replace(value)
}

func isFailureDisposition(disposition string) bool {
	switch disposition {
	case "discard", "revert", "failed", "failure", "blocked", "stalled", "dead_end":
		return true
	default:
		return false
	}
}

func scaffoldAutoresearchWorkspace(workdir string) error {
	artifacts := newAutoresearchArtifacts(workdir, time.Now())
	if err := os.MkdirAll(artifacts.TemplateDir, 0o755); err != nil {
		return err
	}

	programPath := filepath.Join(workdir, defaultProgramFilename)
	planningPath := filepath.Join(workdir, planningFilename)
	files := []struct {
		path    string
		content string
	}{
		{path: programPath, content: defaultProgramScaffold},
		{path: planningPath, content: defaultPlanningScaffold},
		{path: filepath.Join(artifacts.TemplateDir, "plan.md"), content: defaultPlanTemplate},
		{path: filepath.Join(artifacts.TemplateDir, "execution.md"), content: defaultExecutionTemplate},
		{path: filepath.Join(artifacts.TemplateDir, "results.md"), content: defaultResultsTemplate},
		{path: filepath.Join(artifacts.TemplateDir, "insights.md"), content: defaultInsightsTemplate},
		{path: artifacts.LatestContextPath, content: "# Latest Context\n\n- No prior runs yet.\n"},
	}
	for _, file := range files {
		if err := writeFileIfMissing(file.path, []byte(strings.TrimSpace(file.content)+"\n"), 0o644); err != nil {
			return err
		}
	}

	if err := writeFileIfMissing(artifacts.ResultsLedgerPath, nil, 0o644); err != nil {
		return err
	}
	return nil
}

func ensureAutoresearchWorkspace(workdir string) (string, error) {
	expected := autoresearchScaffoldPaths(workdir)
	existing := make([]string, 0, len(expected))
	missing := make([]string, 0, len(expected))
	for _, path := range expected {
		if _, err := os.Stat(path); err == nil {
			existing = append(existing, path)
			continue
		} else if errors.Is(err, fs.ErrNotExist) {
			missing = append(missing, path)
			continue
		} else {
			return "", err
		}
	}

	if len(missing) == 0 {
		return "", nil
	}

	warning := ""
	if len(existing) > 0 {
		relMissing := make([]string, 0, len(missing))
		for _, path := range missing {
			rel, err := filepath.Rel(workdir, path)
			if err != nil {
				rel = path
			}
			relMissing = append(relMissing, rel)
		}
		warning = fmt.Sprintf("warning: autoresearch scaffold is partially present in %s; preserving existing files and creating missing scaffold files: %s", workdir, strings.Join(relMissing, ", "))
	}

	if err := scaffoldAutoresearchWorkspace(workdir); err != nil {
		return "", err
	}
	return warning, nil
}

func autoresearchScaffoldPaths(workdir string) []string {
	artifacts := newAutoresearchArtifacts(workdir, time.Unix(0, 0).UTC())
	return []string{
		filepath.Join(workdir, defaultProgramFilename),
		filepath.Join(workdir, planningFilename),
		filepath.Join(artifacts.TemplateDir, "plan.md"),
		filepath.Join(artifacts.TemplateDir, "execution.md"),
		filepath.Join(artifacts.TemplateDir, "results.md"),
		filepath.Join(artifacts.TemplateDir, "insights.md"),
		artifacts.ResultsLedgerPath,
		artifacts.LatestContextPath,
	}
}

const defaultProgramScaffold = `
# Program

Objective: Replace this with one concrete goal.
Primary evaluator: Replace this with one command or manual validation step.
Prompt mode: autoresearch
Council after failures: 3
Checkpoint commits: false

## Constraints

- Keep changes bounded.
- Prefer one hypothesis per cycle.

## Notes

- Add background, known failures, and guardrails here.
`

const defaultPlanningScaffold = `
# Planning

## Objective

- Keep the current objective aligned with program.md.

## Evaluator

- Reuse one primary evaluator until the hypothesis changes.

## Open Questions

- Capture blockers or questions the next loop should resolve.
`

const defaultPlanTemplate = `
# Plan

- Run: ` + "`{{RUN_ID}}`" + `
- Prompt source: ` + "`{{PROMPT_SOURCE}}`" + ` (` + "`{{PROMPT_SOURCE_PATH}}`" + `)
- Objective: {{OBJECTIVE}}
- Primary evaluator: ` + "`{{PRIMARY_EVALUATOR}}`" + `
- Prompt mode: ` + "`{{PROMPT_MODE}}`" + `
- Council policy: ` + "`{{COUNCIL_POLICY}}`" + `
- Council after failures: {{COUNCIL_AFTER_FAILURES}}
- Checkpoint commits: {{CHECKPOINT_COMMITS}}

## Hypothesis

- {{HYPOTHESIS_PLACEHOLDER}}

## Steps

1. Establish the current baseline.
2. Pick one bounded hypothesis.
3. Make one bounded change.
4. Run the primary evaluator exactly once.
5. Record the result and choose keep, discard, or revert.
`

const defaultExecutionTemplate = `
# Execution

## Actions

- Run directory: ` + "`{{RUN_DIR}}`" + `
- Latest context: ` + "`{{LATEST_CONTEXT_PATH}}`" + `
- Results ledger: ` + "`{{RESULTS_LEDGER_PATH}}`" + `

## Commands And Notes
`

const defaultResultsTemplate = `
# Results

- Status: pending
- Council policy: {{COUNCIL_POLICY}}
- Council triggered at start: {{COUNCIL_TRIGGERED}}
- Primary evaluator: ` + "`{{PRIMARY_EVALUATOR}}`" + `

## Observable Signals

- Fill in the evaluator output or manual validation outcome.

## Disposition

- Choose one: keep, discard, revert, blocked.
`

const defaultInsightsTemplate = `
# Insights

## What Worked

- 

## What Failed

- 

## Avoid Next Time

- 

## Promising Next Directions

- 
`
