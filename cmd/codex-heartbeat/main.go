package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	_ "embed"

	"github.com/creack/pty"
	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

//go:embed heartbeat.md
var defaultPrompt string

const (
	startupHeartbeatDelay = 5 * time.Second
	sessionScanTimeout    = 45 * time.Second
	sessionScanInterval   = 2 * time.Second
	terminateGracePeriod  = 5 * time.Second
)

var errWorkspaceLocked = errors.New("workspace is already locked")

type sharedOptions struct {
	workdir          string
	promptPath       string
	safe             bool
	skipGitRepoCheck bool
}

type workspaceConfig struct {
	Workdir    string
	ProjectDir string
	StatePath  string
	LockPath   string
	LogsDir    string
}

type workspaceState struct {
	Workdir   string    `json:"workdir"`
	SessionID string    `json:"session_id,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type logEvent struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Message   string `json:"message,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	Command   string `json:"command,omitempty"`
	ExitCode  *int   `json:"exit_code,omitempty"`
}

type sessionMetaRecord struct {
	ID        string
	Cwd       string
	Timestamp time.Time
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printRootUsage(os.Stderr)
		return 2
	}

	var err error
	switch args[0] {
	case "pulse", "bootstrap":
		err = runPulseCommand(args[1:])
	case "daemon":
		err = runDaemonCommand(args[1:])
	case "run":
		err = runInteractiveCommand(args[1:])
	case "status":
		err = runStatusCommand(args[1:])
	case "-h", "--help", "help":
		printRootUsage(os.Stdout)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		printRootUsage(os.Stderr)
		return 2
	}

	if err == nil {
		return 0
	}

	fmt.Fprintln(os.Stderr, err)
	if code := exitCodeFromError(err); code >= 0 {
		return code
	}
	return 1
}

func runPulseCommand(args []string) error {
	var opts sharedOptions
	fs := flag.NewFlagSet("pulse", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	registerSharedFlags(fs, &opts)
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: codex-heartbeat pulse --workdir DIR [--prompt FILE] [--safe]")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("pulse does not accept positional arguments")
	}

	cfg, promptText, state, err := prepareWorkspace(opts)
	if err != nil {
		return err
	}

	lock, err := acquireWorkspaceLock(cfg.LockPath)
	if err != nil {
		if errors.Is(err, errWorkspaceLocked) {
			fmt.Fprintln(os.Stderr, "workspace is already locked; skipping pulse")
			return nil
		}
		return err
	}
	defer lock.Close()

	return executePulse(cfg, &state, promptText, opts)
}

func runDaemonCommand(args []string) error {
	var opts sharedOptions
	var interval durationFlag
	var endIn durationFlag

	fs := flag.NewFlagSet("daemon", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	registerSharedFlags(fs, &opts)
	fs.Var(&interval, "interval", "Heartbeat interval (examples: 15m, 2 hours, 1 day)")
	fs.Var(&endIn, "end-in", "Stop the heartbeat after this long (examples: 30m, 2 hours, 1 day)")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: codex-heartbeat daemon --workdir DIR --interval 5m [--prompt FILE] [--end-in 1 day]")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("daemon does not accept positional arguments")
	}
	if !interval.IsSet() {
		return fmt.Errorf("--interval is required for daemon")
	}

	cfg, promptText, state, err := prepareWorkspace(opts)
	if err != nil {
		return err
	}

	lock, err := acquireWorkspaceLock(cfg.LockPath)
	if err != nil {
		return err
	}
	defer lock.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var deadline time.Time
	if endIn.IsSet() {
		deadline = time.Now().Add(endIn.Duration())
	}

	appendEvent(cfg.LogsDir, logEvent{
		Timestamp: time.Now().Format(time.RFC3339),
		Type:      "daemon_start",
		SessionID: state.SessionID,
		Message:   fmt.Sprintf("interval=%s end_in=%s", interval.String(), endIn.String()),
	})

	for {
		if !deadline.IsZero() && time.Now().After(deadline) {
			appendEvent(cfg.LogsDir, logEvent{
				Timestamp: time.Now().Format(time.RFC3339),
				Type:      "daemon_stop",
				SessionID: state.SessionID,
				Message:   "deadline reached",
			})
			return nil
		}

		if err := executePulse(cfg, &state, promptText, opts); err != nil {
			return err
		}

		waitFor := interval.Duration()
		if !deadline.IsZero() {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				appendEvent(cfg.LogsDir, logEvent{
					Timestamp: time.Now().Format(time.RFC3339),
					Type:      "daemon_stop",
					SessionID: state.SessionID,
					Message:   "deadline reached",
				})
				return nil
			}
			if remaining < waitFor {
				waitFor = remaining
			}
		}

		timer := time.NewTimer(waitFor)
		select {
		case <-ctx.Done():
			timer.Stop()
			appendEvent(cfg.LogsDir, logEvent{
				Timestamp: time.Now().Format(time.RFC3339),
				Type:      "daemon_stop",
				SessionID: state.SessionID,
				Message:   ctx.Err().Error(),
			})
			return nil
		case <-timer.C:
		}
	}
}

func runInteractiveCommand(args []string) error {
	var opts sharedOptions
	var interval durationFlag
	var endIn durationFlag

	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	registerSharedFlags(fs, &opts)
	fs.Var(&interval, "interval", "Heartbeat interval (examples: 15m, 2 hours, 1 day)")
	fs.Var(&endIn, "end-in", "Stop the heartbeat after this long (examples: 30m, 2 hours, 1 day)")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: codex-heartbeat run --workdir DIR [--prompt FILE] [--interval 15m] [--end-in 1 day]")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("run does not accept positional arguments")
	}

	cfg, promptText, state, err := prepareWorkspace(opts)
	if err != nil {
		return err
	}

	lock, err := acquireWorkspaceLock(cfg.LockPath)
	if err != nil {
		return err
	}
	defer lock.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if endIn.IsSet() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, endIn.Duration())
		defer cancel()
	}

	runLogPath := filepath.Join(cfg.LogsDir, time.Now().Format("2006-01-02")+"-run.log")
	runLogFile, err := os.OpenFile(runLogPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open run log: %w", err)
	}
	defer runLogFile.Close()

	argsForCodex := buildInteractiveArgs(cfg.Workdir, promptText, state.SessionID, opts.safe)
	cmd := exec.Command("codex", argsForCodex...)
	cmd.Dir = cfg.Workdir
	cmd.Env = os.Environ()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	startedAt := time.Now()
	appendEvent(cfg.LogsDir, logEvent{
		Timestamp: startedAt.Format(time.RFC3339),
		Type:      "run_start",
		SessionID: state.SessionID,
		Command:   "codex " + strings.Join(argsForCodex, " "),
		Message:   fmt.Sprintf("interval=%s end_in=%s", interval.String(), endIn.String()),
	})

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("start codex: %w", err)
	}
	defer ptmx.Close()

	var restore func()
	if term.IsTerminal(int(os.Stdin.Fd())) {
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return fmt.Errorf("set terminal raw mode: %w", err)
		}
		restore = func() {
			_ = term.Restore(int(os.Stdin.Fd()), oldState)
		}
		defer restore()

		if err := pty.InheritSize(os.Stdin, ptmx); err == nil {
			resizeSignals := make(chan os.Signal, 1)
			signal.Notify(resizeSignals, syscall.SIGWINCH)
			defer signal.Stop(resizeSignals)

			go func() {
				for range resizeSignals {
					_ = pty.InheritSize(os.Stdin, ptmx)
				}
			}()
			resizeSignals <- syscall.SIGWINCH
		}
	}

	go trackSessionID(ctx, cfg, &state, startedAt)
	if interval.IsSet() {
		go injectHeartbeatLoop(ctx, ptmx, promptText, interval.Duration(), state.SessionID != "", cfg, &state)
	}

	outputDone := make(chan error, 1)
	go func() {
		_, err := io.Copy(io.MultiWriter(os.Stdout, runLogFile), ptmx)
		if isIgnorableCopyError(err) {
			err = nil
		}
		outputDone <- err
	}()

	if term.IsTerminal(int(os.Stdin.Fd())) {
		go func() {
			_, _ = io.Copy(ptmx, os.Stdin)
		}()
	}

	waitDone := make(chan error, 1)
	go func() {
		waitDone <- cmd.Wait()
	}()

	select {
	case err := <-waitDone:
		copyErr := <-outputDone
		if copyErr != nil {
			return copyErr
		}
		if err != nil {
			appendEvent(cfg.LogsDir, logEvent{
				Timestamp: time.Now().Format(time.RFC3339),
				Type:      "run_stop",
				SessionID: state.SessionID,
				Message:   err.Error(),
			})
			return err
		}
		appendEvent(cfg.LogsDir, logEvent{
			Timestamp: time.Now().Format(time.RFC3339),
			Type:      "run_stop",
			SessionID: state.SessionID,
			Message:   "completed",
		})
		return nil
	case <-ctx.Done():
		appendEvent(cfg.LogsDir, logEvent{
			Timestamp: time.Now().Format(time.RFC3339),
			Type:      "run_stop",
			SessionID: state.SessionID,
			Message:   ctx.Err().Error(),
		})
		signalInteractiveProcess(cmd.Process, syscall.SIGINT)
		var err error
		select {
		case err = <-waitDone:
		case <-time.After(terminateGracePeriod):
			signalInteractiveProcess(cmd.Process, syscall.SIGKILL)
			err = <-waitDone
		}
		<-outputDone
		if err != nil && !isSignalExit(err) {
			return err
		}
		return nil
	}
}

func runStatusCommand(args []string) error {
	var opts sharedOptions
	flagSet := flag.NewFlagSet("status", flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)
	registerSharedFlags(flagSet, &opts)
	flagSet.Usage = func() {
		fmt.Fprintln(flagSet.Output(), "Usage: codex-heartbeat status --workdir DIR")
		flagSet.PrintDefaults()
	}

	if err := flagSet.Parse(args); err != nil {
		return err
	}
	if flagSet.NArg() != 0 {
		return fmt.Errorf("status does not accept positional arguments")
	}
	if strings.TrimSpace(opts.workdir) == "" {
		return fmt.Errorf("--workdir is required")
	}

	cfg, err := newWorkspaceConfig(opts.workdir)
	if err != nil {
		return err
	}

	state, err := loadState(cfg.StatePath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if errors.Is(err, fs.ErrNotExist) {
		state.Workdir = cfg.Workdir
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(state)
}

func registerSharedFlags(fs *flag.FlagSet, opts *sharedOptions) {
	fs.StringVar(&opts.workdir, "workdir", "", "Workspace directory to manage")
	fs.StringVar(&opts.promptPath, "prompt", "", "Optional heartbeat prompt file; defaults to the embedded heartbeat.md")
	fs.BoolVar(&opts.safe, "safe", false, "Do not pass --dangerously-bypass-approvals-and-sandbox to child Codex runs")
	fs.BoolVar(&opts.skipGitRepoCheck, "skip-git-repo-check", true, "Pass --skip-git-repo-check to non-interactive child Codex runs")
}

func prepareWorkspace(opts sharedOptions) (workspaceConfig, string, workspaceState, error) {
	if strings.TrimSpace(opts.workdir) == "" {
		return workspaceConfig{}, "", workspaceState{}, fmt.Errorf("--workdir is required")
	}

	cfg, err := newWorkspaceConfig(opts.workdir)
	if err != nil {
		return workspaceConfig{}, "", workspaceState{}, err
	}
	if err := os.MkdirAll(cfg.LogsDir, 0o755); err != nil {
		return workspaceConfig{}, "", workspaceState{}, fmt.Errorf("create runtime dirs: %w", err)
	}

	promptText, err := loadPrompt(opts.promptPath)
	if err != nil {
		return workspaceConfig{}, "", workspaceState{}, err
	}

	state, err := loadState(cfg.StatePath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return workspaceConfig{}, "", workspaceState{}, err
	}
	if errors.Is(err, fs.ErrNotExist) {
		state = workspaceState{Workdir: cfg.Workdir}
	}
	if state.Workdir == "" {
		state.Workdir = cfg.Workdir
	}

	return cfg, promptText, state, nil
}

func newWorkspaceConfig(workdir string) (workspaceConfig, error) {
	absWorkdir, err := filepath.Abs(workdir)
	if err != nil {
		return workspaceConfig{}, fmt.Errorf("resolve workdir: %w", err)
	}
	info, err := os.Stat(absWorkdir)
	if err != nil {
		return workspaceConfig{}, fmt.Errorf("stat workdir: %w", err)
	}
	if !info.IsDir() {
		return workspaceConfig{}, fmt.Errorf("workdir %q is not a directory", absWorkdir)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return workspaceConfig{}, fmt.Errorf("resolve home dir: %w", err)
	}

	key := workspaceKey(absWorkdir)
	projectDir := filepath.Join(home, ".codex-heartbeat", "projects", key)
	return workspaceConfig{
		Workdir:    absWorkdir,
		ProjectDir: projectDir,
		StatePath:  filepath.Join(projectDir, "state.json"),
		LockPath:   filepath.Join(projectDir, "heartbeat.lock"),
		LogsDir:    filepath.Join(projectDir, "logs"),
	}, nil
}

func workspaceKey(workdir string) string {
	base := filepath.Base(workdir)
	base = strings.ToLower(base)
	base = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-' || r == '_':
			return r
		default:
			return '-'
		}
	}, base)
	base = strings.Trim(base, "-")
	if base == "" {
		base = "workspace"
	}

	sum := sha256.Sum256([]byte(workdir))
	return fmt.Sprintf("%s-%s", base, hex.EncodeToString(sum[:6]))
}

type workspaceLock struct {
	file *os.File
}

func acquireWorkspaceLock(path string) (*workspaceLock, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create lock dir: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open lock file: %w", err)
	}

	if err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		file.Close()
		if errors.Is(err, unix.EWOULDBLOCK) {
			return nil, fmt.Errorf("%w (%s)", errWorkspaceLocked, path)
		}
		return nil, fmt.Errorf("lock workspace: %w", err)
	}

	if err := file.Truncate(0); err == nil {
		_, _ = file.WriteString(fmt.Sprintf("%d\n", os.Getpid()))
	}

	return &workspaceLock{file: file}, nil
}

func (l *workspaceLock) Close() {
	if l == nil || l.file == nil {
		return
	}
	_ = unix.Flock(int(l.file.Fd()), unix.LOCK_UN)
	_ = l.file.Close()
}

func loadPrompt(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return strings.TrimSpace(defaultPrompt), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read prompt file: %w", err)
	}
	prompt := strings.TrimSpace(string(data))
	if prompt == "" {
		return "", fmt.Errorf("prompt file %q is empty", path)
	}
	return prompt, nil
}

func loadState(path string) (workspaceState, error) {
	var state workspaceState
	data, err := os.ReadFile(path)
	if err != nil {
		return state, err
	}
	if err := json.Unmarshal(data, &state); err != nil {
		return state, fmt.Errorf("parse state file: %w", err)
	}
	return state, nil
}

func saveState(path string, state workspaceState) error {
	state.UpdatedAt = time.Now()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state file: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}
	return nil
}

func buildExecArgs(workdir, promptText, sessionID string, opts sharedOptions) []string {
	args := []string{"exec"}
	if !opts.safe {
		args = append(args, "--dangerously-bypass-approvals-and-sandbox")
	}
	args = append(args, "--cd", workdir)
	if opts.skipGitRepoCheck {
		args = append(args, "--skip-git-repo-check")
	}
	if sessionID != "" {
		args = append(args, "resume", sessionID, promptText)
		return args
	}
	return append(args, promptText)
}

func buildInteractiveArgs(workdir, promptText, sessionID string, safe bool) []string {
	args := []string{}
	if !safe {
		args = append(args, "--dangerously-bypass-approvals-and-sandbox")
	}
	args = append(args, "--cd", workdir)
	if sessionID != "" {
		args = append(args, "resume", sessionID)
		return args
	}
	return append(args, promptText)
}

func executePulse(cfg workspaceConfig, state *workspaceState, promptText string, opts sharedOptions) error {
	args := buildExecArgs(cfg.Workdir, promptText, state.SessionID, opts)
	cmd := exec.Command("codex", args...)
	cmd.Dir = cfg.Workdir
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	startedAt := time.Now()
	appendEvent(cfg.LogsDir, logEvent{
		Timestamp: startedAt.Format(time.RFC3339),
		Type:      "pulse_start",
		SessionID: state.SessionID,
		Command:   "codex " + strings.Join(args, " "),
	})

	err := cmd.Run()
	exitCode := exitCodeFromError(err)
	appendEvent(cfg.LogsDir, logEvent{
		Timestamp: time.Now().Format(time.RFC3339),
		Type:      "pulse_stop",
		SessionID: state.SessionID,
		Command:   "codex " + strings.Join(args, " "),
		ExitCode:  intPointer(exitCode),
	})

	if refreshErr := refreshSessionID(cfg, state, startedAt); refreshErr != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", refreshErr)
	}

	return err
}

func trackSessionID(ctx context.Context, cfg workspaceConfig, state *workspaceState, startedAt time.Time) {
	if strings.TrimSpace(state.SessionID) != "" {
		return
	}

	timeout := time.NewTimer(sessionScanTimeout)
	ticker := time.NewTicker(sessionScanInterval)
	defer timeout.Stop()
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timeout.C:
			return
		case <-ticker.C:
			if err := refreshSessionID(cfg, state, startedAt); err == nil && state.SessionID != "" {
				return
			}
		}
	}
}

func refreshSessionID(cfg workspaceConfig, state *workspaceState, startedAt time.Time) error {
	sessionID, _, err := discoverNewestSessionID(cfg.Workdir, startedAt)
	if err != nil {
		return err
	}
	if sessionID == "" || sessionID == state.SessionID {
		return nil
	}

	state.Workdir = cfg.Workdir
	state.SessionID = sessionID
	if err := saveState(cfg.StatePath, *state); err != nil {
		return err
	}
	appendEvent(cfg.LogsDir, logEvent{
		Timestamp: time.Now().Format(time.RFC3339),
		Type:      "session_discovered",
		SessionID: sessionID,
		Message:   "state updated",
	})
	return nil
}

func discoverNewestSessionID(workdir string, notBefore time.Time) (string, time.Time, error) {
	sessionRoot := sessionRootDir()
	var newest sessionMetaRecord

	err := filepath.WalkDir(sessionRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			return nil
		}

		record, err := readSessionMeta(path)
		if err != nil {
			return nil
		}
		if record.ID == "" || !samePath(record.Cwd, workdir) {
			return nil
		}
		if !notBefore.IsZero() && record.Timestamp.Before(notBefore.Add(-2*time.Second)) {
			return nil
		}
		if newest.ID == "" || record.Timestamp.After(newest.Timestamp) {
			newest = record
		}
		return nil
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("scan sessions: %w", err)
	}
	if newest.ID == "" {
		return "", time.Time{}, fmt.Errorf("no session metadata found for %s", workdir)
	}
	return newest.ID, newest.Timestamp, nil
}

func readSessionMeta(path string) (sessionMetaRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return sessionMetaRecord{}, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for i := 0; i < 5; i++ {
		line, err := reader.ReadBytes('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return sessionMetaRecord{}, err
		}
		if len(line) == 0 {
			break
		}

		var envelope struct {
			Type      string `json:"type"`
			Timestamp string `json:"timestamp"`
			Payload   struct {
				ID        string `json:"id"`
				Cwd       string `json:"cwd"`
				Timestamp string `json:"timestamp"`
			} `json:"payload"`
		}
		if jsonErr := json.Unmarshal(line, &envelope); jsonErr == nil && envelope.Type == "session_meta" {
			ts := envelope.Payload.Timestamp
			if ts == "" {
				ts = envelope.Timestamp
			}

			parsedTime, _ := time.Parse(time.RFC3339Nano, ts)
			return sessionMetaRecord{
				ID:        envelope.Payload.ID,
				Cwd:       envelope.Payload.Cwd,
				Timestamp: parsedTime,
			}, nil
		}
		if errors.Is(err, io.EOF) {
			break
		}
	}

	return sessionMetaRecord{}, fmt.Errorf("no session_meta record in %s", path)
}

func sessionRootDir() string {
	if codexHome := strings.TrimSpace(os.Getenv("CODEX_HOME")); codexHome != "" {
		return filepath.Join(codexHome, "sessions")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".codex/sessions"
	}
	return filepath.Join(home, ".codex", "sessions")
}

func samePath(left, right string) bool {
	leftAbs, leftErr := filepath.Abs(left)
	rightAbs, rightErr := filepath.Abs(right)
	if leftErr == nil && rightErr == nil && filepath.Clean(leftAbs) == filepath.Clean(rightAbs) {
		return true
	}

	leftEval, leftErr := filepath.EvalSymlinks(left)
	rightEval, rightErr := filepath.EvalSymlinks(right)
	if leftErr == nil && rightErr == nil && filepath.Clean(leftEval) == filepath.Clean(rightEval) {
		return true
	}

	return false
}

func injectHeartbeatLoop(ctx context.Context, writer io.Writer, promptText string, interval time.Duration, immediate bool, cfg workspaceConfig, state *workspaceState) {
	if interval <= 0 || strings.TrimSpace(promptText) == "" {
		return
	}

	nextDelay := interval
	if immediate {
		nextDelay = startupHeartbeatDelay
	}

	timer := time.NewTimer(nextDelay)
	defer timer.Stop()

	first := true
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if err := injectPrompt(writer, promptText); err == nil {
				appendEvent(cfg.LogsDir, logEvent{
					Timestamp: time.Now().Format(time.RFC3339),
					Type:      "heartbeat_injected",
					SessionID: state.SessionID,
					Message:   formatFlexibleDuration(interval),
				})
			}

			if first && immediate {
				first = false
				timer.Reset(interval)
				continue
			}

			first = false
			timer.Reset(interval)
		}
	}
}

func injectPrompt(writer io.Writer, promptText string) error {
	normalized := strings.ReplaceAll(promptText, "\r\n", "\n")
	_, err := io.WriteString(writer, "\x1b[200~"+normalized+"\x1b[201~\r")
	return err
}

func signalInteractiveProcess(proc *os.Process, sig syscall.Signal) {
	if proc == nil {
		return
	}

	target := proc.Pid
	if pgid, err := syscall.Getpgid(proc.Pid); err == nil && pgid > 0 {
		target = -pgid
	}

	_ = syscall.Kill(target, sig)
}

func appendEvent(logsDir string, event logEvent) {
	if logsDir == "" {
		return
	}
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return
	}

	path := filepath.Join(logsDir, time.Now().Format("2006-01-02")+".jsonl")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	_ = enc.Encode(event)
}

func printRootUsage(w io.Writer) {
	fmt.Fprintln(w, "codex-heartbeat wraps the Codex CLI and can inject heartbeat prompts on a schedule.")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  codex-heartbeat pulse --workdir DIR [--prompt FILE]")
	fmt.Fprintln(w, "  codex-heartbeat run --workdir DIR [--prompt FILE] [--interval 15m] [--end-in 1 day]")
	fmt.Fprintln(w, "  codex-heartbeat daemon --workdir DIR --interval 5m [--end-in 2 hours]")
	fmt.Fprintln(w, "  codex-heartbeat status --workdir DIR")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "The --interval and --end-in flags accept minute, hour, and day units in short or long form.")
	fmt.Fprintln(w, "Examples: 30m, 2h, 1d, 15 minutes, 2 hours, 1 day")
}

func exitCodeFromError(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func intPointer(value int) *int {
	if value < 0 {
		return nil
	}
	return &value
}

func isSignalExit(err error) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}
	status, ok := exitErr.Sys().(syscall.WaitStatus)
	return ok && status.Signaled()
}

func isIgnorableCopyError(err error) bool {
	if err == nil {
		return true
	}
	if strings.Contains(err.Error(), "input/output error") {
		return true
	}
	return false
}
