package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	screenIdlePollInterval = 5 * time.Second
	screenIdlePollCount    = 3
	screenIdleRecentLines  = 8
	screenIdleQuietWindow  = 20 * time.Second
	screenIdleOutputWindow = 30 * time.Second
	screenIdlePromptGrace  = 45 * time.Second
	screenIdleFallbackWait = 60 * time.Minute
	screenSnapshotLimit    = 240
)

var screenElapsedPattern = regexp.MustCompile(`\((?:\d+h )?(?:\d+m )?\d+s\b`)

type screenState int

const (
	screenStateAmbiguous screenState = iota
	screenStateIdle
	screenStateWorking
)

type terminalScreen struct {
	mu sync.RWMutex

	width  int
	height int
	cells  [][]rune

	row int
	col int

	savedRow int
	savedCol int

	parseState screenParseState
	csiBuffer  []byte
	oscEscape  bool
	utf8Buffer []byte
}

type screenParseState int

const (
	screenParseNormal screenParseState = iota
	screenParseEscape
	screenParseCSI
	screenParseOSC
)

type userInputTracker struct {
	mu          sync.RWMutex
	lastInputAt time.Time
}

type outputActivityTracker struct {
	mu           sync.RWMutex
	lastOutputAt time.Time
}

type promptInjectionTracker struct {
	mu           sync.RWMutex
	lastPromptAt time.Time
}

type inputActivityWriter struct {
	tracker *userInputTracker
}

type outputActivityWriter struct {
	tracker *outputActivityTracker
}

type screenPollDecision struct {
	nextIdlePolls int
	shouldInject  bool
	reason        string
}

type screenRuntimeState struct {
	UpdatedAt     time.Time `json:"updated_at"`
	SessionID     string    `json:"session_id,omitempty"`
	Scheduler     string    `json:"scheduler"`
	ScreenState   string    `json:"screen_state"`
	InputQuiet    bool      `json:"input_quiet"`
	OutputQuiet   bool      `json:"output_quiet"`
	Quiet         bool      `json:"quiet"`
	IdlePolls     int       `json:"idle_polls"`
	Reason        string    `json:"reason"`
	LastCheckedAt time.Time `json:"last_checked_at"`
	LastPromptAt  time.Time `json:"last_prompt_at,omitempty"`
	ShouldInject  bool      `json:"should_inject"`
	Injected      bool      `json:"injected"`
	Snapshot      string    `json:"snapshot,omitempty"`
}

type screenPollRecord struct {
	Timestamp    string `json:"timestamp"`
	SessionID    string `json:"session_id,omitempty"`
	Scheduler    string `json:"scheduler"`
	ScreenState  string `json:"screen_state"`
	InputQuiet   bool   `json:"input_quiet"`
	OutputQuiet  bool   `json:"output_quiet"`
	Quiet        bool   `json:"quiet"`
	IdlePolls    int    `json:"idle_polls"`
	Reason       string `json:"reason"`
	LastPromptAt string `json:"last_prompt_at,omitempty"`
	ShouldInject bool   `json:"should_inject"`
	Injected     bool   `json:"injected"`
	Snapshot     string `json:"snapshot,omitempty"`
}

func newTerminalScreen(width, height int) *terminalScreen {
	if width <= 0 {
		width = 120
	}
	if height <= 0 {
		height = 40
	}

	screen := &terminalScreen{
		width:  width,
		height: height,
		cells:  make([][]rune, height),
	}
	for i := range screen.cells {
		screen.cells[i] = blankScreenLine(width)
	}
	return screen
}

func blankScreenLine(width int) []rune {
	line := make([]rune, width)
	for i := range line {
		line[i] = ' '
	}
	return line
}

func (s *terminalScreen) Resize(width, height int) {
	if width <= 0 || height <= 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if width == s.width && height == s.height {
		return
	}

	next := make([][]rune, height)
	for i := range next {
		next[i] = blankScreenLine(width)
	}

	copyRows := min(height, s.height)
	copyCols := min(width, s.width)
	for row := 0; row < copyRows; row++ {
		copy(next[row][:copyCols], s.cells[row][:copyCols])
	}

	s.width = width
	s.height = height
	s.cells = next
	s.row = clampInt(s.row, 0, s.height-1)
	s.col = clampInt(s.col, 0, s.width-1)
	s.savedRow = clampInt(s.savedRow, 0, s.height-1)
	s.savedCol = clampInt(s.savedCol, 0, s.width-1)
}

func (s *terminalScreen) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := 0; i < len(p); i++ {
		b := p[i]
		switch s.parseState {
		case screenParseEscape:
			switch b {
			case '[':
				s.parseState = screenParseCSI
				s.csiBuffer = s.csiBuffer[:0]
			case ']':
				s.parseState = screenParseOSC
				s.oscEscape = false
			case '7':
				s.savedRow = s.row
				s.savedCol = s.col
				s.parseState = screenParseNormal
			case '8':
				s.row = s.savedRow
				s.col = s.savedCol
				s.parseState = screenParseNormal
			case 'c':
				s.clearAll()
				s.row = 0
				s.col = 0
				s.parseState = screenParseNormal
			default:
				s.parseState = screenParseNormal
			}
			continue
		case screenParseCSI:
			if b >= 0x40 && b <= 0x7e {
				s.handleCSI(string(s.csiBuffer), b)
				s.csiBuffer = s.csiBuffer[:0]
				s.parseState = screenParseNormal
				continue
			}
			s.csiBuffer = append(s.csiBuffer, b)
			continue
		case screenParseOSC:
			if s.oscEscape {
				s.oscEscape = false
				if b == '\\' {
					s.parseState = screenParseNormal
				}
				continue
			}
			if b == '\a' {
				s.parseState = screenParseNormal
				continue
			}
			if b == 0x1b {
				s.oscEscape = true
			}
			continue
		}

		switch {
		case b == 0x1b:
			s.flushUTF8Buffer()
			s.parseState = screenParseEscape
		case b == '\r':
			s.flushUTF8Buffer()
			s.col = 0
		case b == '\n':
			s.flushUTF8Buffer()
			s.newLine()
		case b == '\b':
			s.flushUTF8Buffer()
			if s.col > 0 {
				s.col--
			}
		case b == '\t':
			s.flushUTF8Buffer()
			nextTab := ((s.col / 8) + 1) * 8
			s.col = min(nextTab, s.width-1)
		case len(s.utf8Buffer) > 0 || b >= utf8.RuneSelf:
			s.utf8Buffer = append(s.utf8Buffer, b)
			s.flushUTF8Buffer()
		case b >= 0x20 && b != 0x7f:
			s.putRune(rune(b))
		}
	}

	return len(p), nil
}

func (s *terminalScreen) Snapshot() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return renderScreenLines(s.cells)
}

func (s *terminalScreen) RecentSnapshot(lineCount int) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if lineCount <= 0 || lineCount >= s.height {
		return renderScreenLines(s.cells)
	}

	end := clampInt(s.row+1, 1, s.height)
	start := max(0, end-lineCount)
	return renderScreenLines(s.cells[start:end])
}

func renderScreenLines(lines [][]rune) string {
	if len(lines) == 0 {
		return ""
	}

	rendered := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimRight(string(line), " ")
		rendered = append(rendered, trimmed)
	}
	return strings.TrimSpace(strings.Join(rendered, "\n"))
}

func (t *userInputTracker) Mark(at time.Time) {
	if t == nil {
		return
	}

	t.mu.Lock()
	t.lastInputAt = at
	t.mu.Unlock()
}

func (t *userInputTracker) IsQuiet(now time.Time, window time.Duration) bool {
	if t == nil || window <= 0 {
		return true
	}

	t.mu.RLock()
	lastInputAt := t.lastInputAt
	t.mu.RUnlock()

	if lastInputAt.IsZero() {
		return true
	}

	return now.Sub(lastInputAt) >= window
}

func (t *outputActivityTracker) Mark(at time.Time) {
	if t == nil {
		return
	}

	t.mu.Lock()
	t.lastOutputAt = at
	t.mu.Unlock()
}

func (t *outputActivityTracker) IsQuiet(now time.Time, window time.Duration) bool {
	if t == nil || window <= 0 {
		return true
	}

	t.mu.RLock()
	lastOutputAt := t.lastOutputAt
	t.mu.RUnlock()

	if lastOutputAt.IsZero() {
		return true
	}

	return now.Sub(lastOutputAt) >= window
}

func newPromptInjectionTracker(initial time.Time) *promptInjectionTracker {
	return &promptInjectionTracker{lastPromptAt: initial}
}

func (t *promptInjectionTracker) Mark(at time.Time) {
	if t == nil {
		return
	}

	t.mu.Lock()
	t.lastPromptAt = at
	t.mu.Unlock()
}

func (t *promptInjectionTracker) LastPromptAt() time.Time {
	if t == nil {
		return time.Time{}
	}

	t.mu.RLock()
	lastPromptAt := t.lastPromptAt
	t.mu.RUnlock()
	return lastPromptAt
}

func (w inputActivityWriter) Write(p []byte) (int, error) {
	if len(p) > 0 && w.tracker != nil {
		w.tracker.Mark(time.Now())
	}
	return len(p), nil
}

func (w outputActivityWriter) Write(p []byte) (int, error) {
	if len(p) > 0 && w.tracker != nil {
		w.tracker.Mark(time.Now())
	}
	return len(p), nil
}

func trackUserInput(reader io.Reader, tracker *userInputTracker) io.Reader {
	if tracker == nil {
		return reader
	}
	return io.TeeReader(reader, inputActivityWriter{tracker: tracker})
}

func screenStateLabel(state screenState) string {
	switch state {
	case screenStateIdle:
		return "idle"
	case screenStateWorking:
		return "working"
	default:
		return "ambiguous"
	}
}

func trimScreenSnapshot(snapshot string) string {
	snapshot = strings.TrimSpace(snapshot)
	if snapshot == "" {
		return ""
	}

	runes := []rune(snapshot)
	if len(runes) <= screenSnapshotLimit {
		return snapshot
	}
	return string(runes[:screenSnapshotLimit]) + "..."
}

func classifyScreenSnapshot(snapshot string) screenState {
	lines := screenSnapshotEvidenceLines(snapshot)
	if len(lines) == 0 {
		return screenStateAmbiguous
	}

	activeScore := 0
	idleScore := 0
	for i, line := range lines {
		weight := screenEvidenceWeight(i, len(lines))
		window := line.normalized
		if i+1 < len(lines) {
			window = window + " " + lines[i+1].normalized
		}

		activeLineScore := 0
		if !screenHasPendingInterruptHint(window) && !screenHasHistoricalActivity(window) {
			hasElapsed := screenElapsedPattern.MatchString(window)
			hasInterrupt := screenHasInterruptHint(window)
			hasHeader := screenHasActiveHeader(window)
			hasStatusLead := screenHasStatusLead(line.raw)
			hasLiveBackground := screenHasLiveBackgroundCue(window)

			switch {
			case hasElapsed && (hasInterrupt || hasHeader):
				activeLineScore = 4
			case hasElapsed && hasStatusLead:
				activeLineScore = 3
			case hasHeader && hasInterrupt:
				activeLineScore = 3
			case hasLiveBackground && (hasElapsed || hasHeader || hasInterrupt):
				activeLineScore = 3
			case hasLiveBackground:
				activeLineScore = 3
			}
		}
		activeScore += activeLineScore * weight

		idleLineScore := 0
		if screenHasIdleMarker(window) {
			idleLineScore = 4
		}
		if screenHasPromptReady(line.raw, line.normalized) {
			idleLineScore = max(idleLineScore, 2)
		}
		idleScore += idleLineScore * weight
	}

	switch {
	case activeScore >= 4 && activeScore > idleScore:
		return screenStateWorking
	case idleScore >= 4 && idleScore >= activeScore:
		return screenStateIdle
	case idleScore >= 2 && activeScore <= 0:
		return screenStateIdle
	default:
		return screenStateAmbiguous
	}
}

type screenEvidenceLine struct {
	raw        string
	normalized string
}

func screenSnapshotEvidenceLines(snapshot string) []screenEvidenceLine {
	rawLines := strings.Split(snapshot, "\n")
	lines := make([]screenEvidenceLine, 0, len(rawLines))
	for _, raw := range rawLines {
		normalized := normalizeScreenSnapshot(raw)
		if normalized == "" {
			continue
		}
		lines = append(lines, screenEvidenceLine{
			raw:        raw,
			normalized: normalized,
		})
	}

	return lines
}

func screenEvidenceWeight(index, total int) int {
	switch {
	case index >= total-2:
		return 3
	case index >= total-screenIdleRecentLines:
		return 2
	default:
		return 1
	}
}

func screenHasInterruptHint(normalized string) bool {
	return strings.Contains(normalized, "esc to interrupt") ||
		strings.Contains(normalized, "esc to …") ||
		strings.Contains(normalized, "esc to ...") ||
		strings.Contains(normalized, "esc to")
}

func screenHasPendingInterruptHint(normalized string) bool {
	return strings.Contains(normalized, "press esc to interrupt and send immediately")
}

func screenHasHistoricalActivity(normalized string) bool {
	historicalPhrases := []string{
		"waited for background terminal",
		"interacted with background terminal",
		"no background terminals running",
		"no background terminal running",
	}
	for _, phrase := range historicalPhrases {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}
	return false
}

func screenHasActiveHeader(normalized string) bool {
	activeHeaders := []string{
		"working",
		"booting mcp server",
		"starting mcp",
		"analyzing",
		"investigating",
		"reviewing",
		"waiting for background terminal",
	}
	for _, phrase := range activeHeaders {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}
	return false
}

func screenHasLiveBackgroundCue(normalized string) bool {
	return strings.Contains(normalized, "background terminal running") ||
		strings.Contains(normalized, "background terminals running") ||
		strings.Contains(normalized, "/ps to view")
}

func screenHasIdleMarker(normalized string) bool {
	idlePhrases := []string{
		"token usage:",
		"context compacted",
		"conversation interrupted - tell the model what to do differently",
		"to continue this session, run codex resume",
	}
	for _, phrase := range idlePhrases {
		if strings.Contains(normalized, phrase) {
			return true
		}
	}
	return screenHasHistoricalActivity(normalized)
}

func screenHasPromptReady(raw, normalized string) bool {
	return strings.Contains(raw, "›") && !strings.Contains(normalized, "loading /model to change")
}

func screenHasStatusLead(raw string) bool {
	trimmed := strings.TrimSpace(raw)
	return strings.HasPrefix(trimmed, "• ") || strings.HasPrefix(trimmed, "◦ ")
}

func normalizeScreenSnapshot(snapshot string) string {
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return -1
		}
		return unicode.ToLower(r)
	}, snapshot)
	return strings.Join(strings.Fields(cleaned), " ")
}

func evaluateScreenIdlePoll(now time.Time, inputQuiet bool, outputQuiet bool, currentState screenState, idlePolls int, lastPromptAt time.Time) screenPollDecision {
	quiet := inputQuiet && outputQuiet
	if screenIdleRecentPrompt(now, lastPromptAt) {
		return screenPollDecision{
			nextIdlePolls: 0,
			reason:        "recent_prompt_grace",
		}
	}
	if screenIdleFallbackDue(now, lastPromptAt, quiet) {
		return screenPollDecision{
			nextIdlePolls: 0,
			shouldInject:  true,
			reason:        "fallback_due",
		}
	}
	switch currentState {
	case screenStateWorking:
		return screenPollDecision{
			nextIdlePolls: 0,
			reason:        "screen_working",
		}
	case screenStateAmbiguous:
		return screenPollDecision{
			nextIdlePolls: 0,
			reason:        "screen_ambiguous",
		}
	default:
		if !outputQuiet {
			return screenPollDecision{
				nextIdlePolls: 0,
				reason:        "idle_blocked_recent_output",
			}
		}
		if !inputQuiet {
			return screenPollDecision{
				nextIdlePolls: 0,
				reason:        "idle_blocked_recent_input",
			}
		}

		nextIdlePolls := min(idlePolls+1, screenIdlePollCount)
		if nextIdlePolls < screenIdlePollCount {
			return screenPollDecision{
				nextIdlePolls: nextIdlePolls,
				reason:        "idle_accumulating",
			}
		}
		return screenPollDecision{
			nextIdlePolls: 0,
			shouldInject:  true,
			reason:        "idle_threshold_reached",
		}
	}
}

func screenIdleRecentPrompt(now, lastPromptAt time.Time) bool {
	if lastPromptAt.IsZero() {
		return false
	}
	return now.Sub(lastPromptAt) < screenIdlePromptGrace
}

func screenIdleFallbackDue(now, lastPromptAt time.Time, quiet bool) bool {
	if !quiet || lastPromptAt.IsZero() {
		return false
	}

	return now.Sub(lastPromptAt) >= screenIdleFallbackWait
}

func screenStateFilePath(projectDir string) string {
	return filepath.Join(projectDir, "screen-state.json")
}

func appendScreenPoll(logsDir string, poll screenPollRecord) error {
	if logsDir == "" {
		return nil
	}
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return err
	}

	logDate := time.Now()
	if parsed, err := time.Parse(time.RFC3339, poll.Timestamp); err == nil {
		logDate = parsed
	}
	path := filepath.Join(logsDir, logDate.Format("2006-01-02")+"-screen.jsonl")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	return enc.Encode(poll)
}

func saveScreenState(path string, state screenRuntimeState) error {
	state.UpdatedAt = time.Now()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func persistScreenDiagnostics(cfg workspaceConfig, state screenRuntimeState, poll screenPollRecord) {
	_ = saveScreenState(screenStateFilePath(cfg.ProjectDir), state)
	_ = appendScreenPoll(cfg.LogsDir, poll)
}

func injectScreenIdleLoop(ctx context.Context, writer io.Writer, prompts promptResolver, artifacts autoresearchArtifacts, screen *terminalScreen, inputTracker *userInputTracker, outputTracker *outputActivityTracker, promptTracker *promptInjectionTracker, cfg workspaceConfig, state *workspaceState, errCh chan<- error) {
	if screen == nil {
		return
	}

	ticker := time.NewTicker(screenIdlePollInterval)
	defer ticker.Stop()
	rolloutInspector := newSessionRolloutInspector()

	persistScreenDiagnostics(cfg, screenRuntimeState{
		SessionID:     state.SessionID,
		Scheduler:     screenIdleHeartbeatSummary(),
		ScreenState:   screenStateLabel(screenStateAmbiguous),
		InputQuiet:    true,
		OutputQuiet:   true,
		Quiet:         true,
		IdlePolls:     0,
		Reason:        "starting",
		LastCheckedAt: time.Now(),
		LastPromptAt:  promptTracker.LastPromptAt(),
	}, screenPollRecord{
		Timestamp:   time.Now().Format(time.RFC3339),
		SessionID:   state.SessionID,
		Scheduler:   screenIdleHeartbeatSummary(),
		ScreenState: screenStateLabel(screenStateAmbiguous),
		InputQuiet:  true,
		OutputQuiet: true,
		Quiet:       true,
		IdlePolls:   0,
		Reason:      "starting",
	})

	idlePolls := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			inputQuiet := inputTracker.IsQuiet(now, screenIdleQuietWindow)
			outputQuiet := outputTracker.IsQuiet(now, screenIdleOutputWindow)
			quiet := inputQuiet && outputQuiet
			lastPromptAt := promptTracker.LastPromptAt()
			snapshot := screen.RecentSnapshot(screenIdleRecentLines)
			currentState := classifyScreenSnapshot(snapshot)
			tiebreakReason := ""
			currentState, tiebreakReason = rolloutInspector.Resolve(currentState, state.SessionID)

			decision := evaluateScreenIdlePoll(now, inputQuiet, outputQuiet, currentState, idlePolls, lastPromptAt)
			if paused, pauseReason, err := heartbeatPauseState(artifacts); err != nil {
				reportAsyncError(errCh, err)
				return
			} else if paused {
				decision.shouldInject = false
				decision.reason = pauseReason
			} else if tiebreakReason != "" {
				decision.reason = decision.reason + ":" + tiebreakReason
			}
			idlePolls = decision.nextIdlePolls

			poll := screenPollRecord{
				Timestamp:    now.Format(time.RFC3339),
				SessionID:    state.SessionID,
				Scheduler:    screenIdleHeartbeatSummary(),
				ScreenState:  screenStateLabel(currentState),
				InputQuiet:   inputQuiet,
				OutputQuiet:  outputQuiet,
				Quiet:        quiet,
				IdlePolls:    idlePolls,
				Reason:       decision.reason,
				LastPromptAt: lastPromptAt.Format(time.RFC3339),
				ShouldInject: decision.shouldInject,
				Snapshot:     trimScreenSnapshot(snapshot),
			}
			runtimeState := screenRuntimeState{
				SessionID:     state.SessionID,
				Scheduler:     screenIdleHeartbeatSummary(),
				ScreenState:   screenStateLabel(currentState),
				InputQuiet:    inputQuiet,
				OutputQuiet:   outputQuiet,
				Quiet:         quiet,
				IdlePolls:     idlePolls,
				Reason:        decision.reason,
				LastCheckedAt: now,
				LastPromptAt:  lastPromptAt,
				ShouldInject:  decision.shouldInject,
				Snapshot:      trimScreenSnapshot(snapshot),
			}

			if decision.shouldInject {
				resolution, err := prompts.Resolve(artifacts)
				if err != nil {
					reportAsyncError(errCh, err)
					return
				}
				if err := injectPrompt(writer, resolution.Text); err == nil {
					promptTracker.Mark(now)
					poll.Injected = true
					poll.LastPromptAt = now.Format(time.RFC3339)
					runtimeState.Injected = true
					runtimeState.LastPromptAt = now
					_ = appendExecutionNote(artifacts.ExecutionPath, fmt.Sprintf("screen-idle heartbeat injected with prompt source `%s`", resolution.Source))
					appendEvent(cfg.LogsDir, logEvent{
						Timestamp: now.Format(time.RFC3339),
						Type:      "heartbeat_injected",
						SessionID: state.SessionID,
						Message:   screenIdleHeartbeatSummary(),
					})
				}
			}

			persistScreenDiagnostics(cfg, runtimeState, poll)
		}
	}
}

func screenIdleHeartbeatSummary() string {
	fallback := fmt.Sprintf("%dm", int(screenIdleFallbackWait/time.Minute))
	return fmt.Sprintf(
		"screen-idle=%dx%s/input-quiet=%s/output-quiet=%s/grace=%s/fallback=%s",
		screenIdlePollCount,
		formatFlexibleDuration(screenIdlePollInterval),
		formatFlexibleDuration(screenIdleQuietWindow),
		formatFlexibleDuration(screenIdleOutputWindow),
		formatFlexibleDuration(screenIdlePromptGrace),
		fallback,
	)
}

func (s *terminalScreen) flushUTF8Buffer() {
	for len(s.utf8Buffer) > 0 {
		if !utf8.FullRune(s.utf8Buffer) {
			return
		}
		r, size := utf8.DecodeRune(s.utf8Buffer)
		if r == utf8.RuneError && size == 1 {
			s.utf8Buffer = s.utf8Buffer[1:]
			continue
		}
		s.utf8Buffer = s.utf8Buffer[size:]
		if unicode.IsControl(r) {
			continue
		}
		s.putRune(r)
	}
}

func (s *terminalScreen) handleCSI(raw string, final byte) {
	switch final {
	case 'A':
		s.row = clampInt(s.row-parseCSIInt(raw, 1), 0, s.height-1)
	case 'B':
		s.row = clampInt(s.row+parseCSIInt(raw, 1), 0, s.height-1)
	case 'C':
		s.col = clampInt(s.col+parseCSIInt(raw, 1), 0, s.width-1)
	case 'D':
		s.col = clampInt(s.col-parseCSIInt(raw, 1), 0, s.width-1)
	case 'E':
		s.row = clampInt(s.row+parseCSIInt(raw, 1), 0, s.height-1)
		s.col = 0
	case 'F':
		s.row = clampInt(s.row-parseCSIInt(raw, 1), 0, s.height-1)
		s.col = 0
	case 'G':
		s.col = clampInt(parseCSIInt(raw, 1)-1, 0, s.width-1)
	case 'H', 'f':
		row, col := parseCSICoordinates(raw)
		s.row = clampInt(row-1, 0, s.height-1)
		s.col = clampInt(col-1, 0, s.width-1)
	case 'J':
		s.clearDisplay(parseCSIInt(raw, 0))
	case 'K':
		s.clearLine(parseCSIInt(raw, 0))
	case 'm', 'r', 'h', 'l':
		return
	case 's':
		s.savedRow = s.row
		s.savedCol = s.col
	case 'u':
		s.row = s.savedRow
		s.col = s.savedCol
	}
}

func (s *terminalScreen) newLine() {
	s.row++
	s.col = 0
	if s.row < s.height {
		return
	}
	s.scrollUp()
	s.row = s.height - 1
}

func (s *terminalScreen) putRune(r rune) {
	if s.width == 0 || s.height == 0 {
		return
	}
	s.cells[s.row][s.col] = r
	s.col++
	if s.col < s.width {
		return
	}
	s.col = 0
	s.row++
	if s.row < s.height {
		return
	}
	s.scrollUp()
	s.row = s.height - 1
}

func (s *terminalScreen) scrollUp() {
	copy(s.cells, s.cells[1:])
	s.cells[s.height-1] = blankScreenLine(s.width)
}

func (s *terminalScreen) clearAll() {
	for row := range s.cells {
		s.cells[row] = blankScreenLine(s.width)
	}
}

func (s *terminalScreen) clearDisplay(mode int) {
	switch mode {
	case 1:
		for row := 0; row < s.row; row++ {
			s.cells[row] = blankScreenLine(s.width)
		}
		for col := 0; col <= s.col; col++ {
			s.cells[s.row][col] = ' '
		}
	case 2:
		s.clearAll()
	default:
		for col := s.col; col < s.width; col++ {
			s.cells[s.row][col] = ' '
		}
		for row := s.row + 1; row < s.height; row++ {
			s.cells[row] = blankScreenLine(s.width)
		}
	}
}

func (s *terminalScreen) clearLine(mode int) {
	switch mode {
	case 1:
		for col := 0; col <= s.col; col++ {
			s.cells[s.row][col] = ' '
		}
	case 2:
		s.cells[s.row] = blankScreenLine(s.width)
	default:
		for col := s.col; col < s.width; col++ {
			s.cells[s.row][col] = ' '
		}
	}
}

func parseCSIInt(raw string, fallback int) int {
	raw = strings.TrimLeft(raw, "?>")
	if raw == "" {
		return fallback
	}
	if idx := strings.IndexByte(raw, ';'); idx >= 0 {
		raw = raw[:idx]
	}
	var value int
	for _, r := range raw {
		if r < '0' || r > '9' {
			return fallback
		}
		value = value*10 + int(r-'0')
	}
	if value == 0 {
		return fallback
	}
	return value
}

func parseCSICoordinates(raw string) (row, col int) {
	row = 1
	col = 1
	raw = strings.TrimLeft(raw, "?>")
	if raw == "" {
		return row, col
	}

	parts := strings.Split(raw, ";")
	if len(parts) > 0 {
		row = parseCSIInt(parts[0], 1)
	}
	if len(parts) > 1 {
		col = parseCSIInt(parts[1], 1)
	}
	return row, col
}

func clampInt(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
