package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	screenIdlePollInterval = 10 * time.Second
	screenIdlePollCount    = 3
	screenIdleRecentLines  = 8
	screenIdleQuietWindow  = screenIdlePollInterval * screenIdlePollCount
	screenIdleFallbackWait = 60 * time.Minute
)

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

type promptInjectionTracker struct {
	mu           sync.RWMutex
	lastPromptAt time.Time
}

type inputActivityWriter struct {
	tracker *userInputTracker
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

func trackUserInput(reader io.Reader, tracker *userInputTracker) io.Reader {
	if tracker == nil {
		return reader
	}
	return io.TeeReader(reader, inputActivityWriter{tracker: tracker})
}

func classifyScreenSnapshot(snapshot string) screenState {
	normalized := normalizeScreenSnapshot(snapshot)
	if normalized == "" {
		return screenStateAmbiguous
	}

	activePhrases := []string{
		"working (",
		"working(",
		"esc to interrupt",
		"booting mcp server",
		"messages to be submitted after next tool call",
		"background terminal running",
	}
	for _, phrase := range activePhrases {
		if strings.Contains(normalized, phrase) {
			return screenStateWorking
		}
	}

	idlePhrases := []string{
		"token usage:",
		"conversation interrupted - tell the model what to do differently",
		"to continue this session, run codex resume",
	}
	for _, phrase := range idlePhrases {
		if strings.Contains(normalized, phrase) {
			return screenStateIdle
		}
	}

	if strings.Contains(snapshot, "›") && !strings.Contains(normalized, "loading /model to change") {
		return screenStateIdle
	}

	return screenStateAmbiguous
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

func screenIdleFallbackDue(now, lastPromptAt time.Time, quiet bool) bool {
	if !quiet || lastPromptAt.IsZero() {
		return false
	}

	return now.Sub(lastPromptAt) >= screenIdleFallbackWait
}

func advanceScreenIdlePolls(idlePolls int, quiet bool, currentState screenState) (nextPolls int, shouldInject bool) {
	if !quiet || currentState != screenStateIdle {
		return 0, false
	}

	idlePolls++
	if idlePolls < screenIdlePollCount {
		return idlePolls, false
	}

	return 0, true
}

func injectScreenIdleLoop(ctx context.Context, writer io.Writer, promptText string, screen *terminalScreen, inputTracker *userInputTracker, promptTracker *promptInjectionTracker, cfg workspaceConfig, state *workspaceState) {
	if strings.TrimSpace(promptText) == "" || screen == nil {
		return
	}

	ticker := time.NewTicker(screenIdlePollInterval)
	defer ticker.Stop()

	idlePolls := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			quiet := inputTracker.IsQuiet(now, screenIdleQuietWindow)
			if screenIdleFallbackDue(now, promptTracker.LastPromptAt(), quiet) {
				if err := injectPrompt(writer, promptText); err == nil {
					promptTracker.Mark(now)
					appendEvent(cfg.LogsDir, logEvent{
						Timestamp: now.Format(time.RFC3339),
						Type:      "heartbeat_injected",
						SessionID: state.SessionID,
						Message:   screenIdleHeartbeatSummary(),
					})
				}
				idlePolls = 0
				continue
			}

			currentState := screenStateAmbiguous
			if quiet {
				currentState = classifyScreenSnapshot(screen.RecentSnapshot(screenIdleRecentLines))
			}

			var shouldInject bool
			idlePolls, shouldInject = advanceScreenIdlePolls(idlePolls, quiet, currentState)
			if !shouldInject {
				continue
			}

			if err := injectPrompt(writer, promptText); err == nil {
				promptTracker.Mark(now)
				appendEvent(cfg.LogsDir, logEvent{
					Timestamp: now.Format(time.RFC3339),
					Type:      "heartbeat_injected",
					SessionID: state.SessionID,
					Message:   screenIdleHeartbeatSummary(),
				})
			}
		}
	}
}

func screenIdleHeartbeatSummary() string {
	fallback := fmt.Sprintf("%dm", int(screenIdleFallbackWait/time.Minute))
	return fmt.Sprintf("screen-idle=%dx%s/fallback=%s", screenIdlePollCount, formatFlexibleDuration(screenIdlePollInterval), fallback)
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
