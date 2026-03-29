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
	"runtime"
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
	workdir    string
	promptPath string
	council    bool
	safe       bool
}

type launchOverrides struct {
	Profile              string `json:"profile,omitempty"`
	Model                string `json:"model,omitempty"`
	ModelReasoningEffort string `json:"model_reasoning_effort,omitempty"`
}

type statusOutput struct {
	workspaceState
	LaunchSettings *launchOverrides   `json:"launch_settings,omitempty"`
	HermesParity   hermesParityStatus `json:"hermes_parity"`
}

type hermesParityStatus struct {
	Equivalent bool     `json:"equivalent"`
	Missing    []string `json:"missing"`
	TaskList   []string `json:"task_list"`
	ClaimRule  string   `json:"claim_rule"`
}

type promptSource struct {
	path      string
	cachePath string
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
	case "run":
		err = runInteractiveCommand(args[1:])
	case "status":
		err = runStatusCommand(args[1:])
	case "-h", "--help", "help":
		printRootUsage(os.Stdout)
		return 0
	default:
		if strings.HasPrefix(args[0], "-") {
			err = runInteractiveCommand(args)
			break
		}
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

func runInteractiveCommand(args []string) error {
	var opts sharedOptions
	var interval durationFlag
	var endIn durationFlag
	var noAltScreen bool
	var altScreen bool
	var screenIdleHeartbeat bool

	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	registerRunFlags(fs, &opts)
	fs.Var(&interval, "interval", "Heartbeat interval (examples: 15m, 2 hours, 1 day)")
	fs.Var(&endIn, "end-in", "Stop the heartbeat after this long (examples: 30m, 2 hours, 1 day)")
	fs.BoolVar(&noAltScreen, "no-alt-screen", false, "Run Codex inline so the wrapper banner stays visible in scrollback")
	fs.BoolVar(&altScreen, "alt-screen", false, "Force Codex to use the alternate screen")
	fs.BoolVar(&screenIdleHeartbeat, "screen-idle-heartbeat", false, "Explicitly select the default screen-aware heartbeat mode (15s idle detection, 20s input quiet gate, and 60m fallback)")
	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage: codex-heartbeat run --workdir DIR [--prompt FILE] [--council] [--interval 15m] [--screen-idle-heartbeat] [--end-in 1 day] [--no-alt-screen]")
		fs.PrintDefaults()
	}

	help, err := parseFlagSet(fs, args)
	if err != nil {
		return err
	}
	if help {
		return nil
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("run does not accept positional arguments")
	}
	if interval.IsSet() && screenIdleHeartbeat {
		return fmt.Errorf("--interval and --screen-idle-heartbeat cannot be used together")
	}
	useScreenIdleHeartbeat := useScreenIdleScheduler(interval, screenIdleHeartbeat)
	useNoAltScreen, err := resolveNoAltScreen(noAltScreen, altScreen)
	if err != nil {
		return err
	}

	cfg, prompts, state, err := prepareWorkspace(opts)
	if err != nil {
		return err
	}

	lock, err := acquireWorkspaceLock(cfg.LockPath)
	if err != nil {
		return err
	}
	defer lock.Close()

	artifacts := newAutoresearchArtifacts(cfg.Workdir, time.Now())
	initialResolution, err := prompts.Resolve(artifacts)
	if err != nil {
		return err
	}
	if err := recordRunStart(artifacts, initialResolution, "run"); err != nil {
		return err
	}

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

	sendPromptOnLaunch, injectImmediately := interactiveLaunchBehavior(state.SessionID)
	launchPromptText := ""
	if sendPromptOnLaunch {
		launchPromptText = initialResolution.Text
	}
	overrides := newLaunchOverrides(initialResolution.Program)
	argsForCodex := buildInteractiveArgs(cfg.Workdir, launchPromptText, state.SessionID, opts.safe, sendPromptOnLaunch, useNoAltScreen, overrides)
	if summary := overrides.Summary(); summary != "" {
		if err := appendExecutionNote(artifacts.ExecutionPath, "launch overrides: "+summary); err != nil {
			return err
		}
	}
	cmd := exec.Command("codex", argsForCodex...)
	cmd.Dir = cfg.Workdir
	cmd.Env = os.Environ()

	startedAt := time.Now()
	promptTracker := newPromptInjectionTracker(startedAt)
	appendEvent(cfg.LogsDir, logEvent{
		Timestamp: startedAt.Format(time.RFC3339),
		Type:      "run_start",
		SessionID: state.SessionID,
		Command:   "codex " + strings.Join(argsForCodex, " "),
		Message:   fmt.Sprintf("heartbeat=%s end_in=%s overrides=%s", runHeartbeatMode(interval, screenIdleHeartbeat), endIn.String(), launchSummaryOrNone(overrides)),
	})

	printRunBanner(cfg, state, interval, endIn, useNoAltScreen, screenIdleHeartbeat)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("start codex: %w", err)
	}
	defer ptmx.Close()

	rows, cols, sizeErr := pty.Getsize(ptmx)
	if sizeErr != nil {
		rows = 40
		cols = 120
	}
	screen := newTerminalScreen(cols, rows)
	var inputTracker *userInputTracker
	if useScreenIdleHeartbeat {
		inputTracker = &userInputTracker{}
	}

	setTerminalTitle(runTerminalTitle(cfg))

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
					if rows, cols, err := pty.Getsize(ptmx); err == nil {
						screen.Resize(cols, rows)
					}
				}
			}()
			resizeSignals <- syscall.SIGWINCH
		}
	}

	go trackSessionID(ctx, cfg, &state, startedAt)
	schedulerErrors := make(chan error, 1)
	if useScreenIdleHeartbeat {
		go injectScreenIdleLoop(ctx, ptmx, prompts, artifacts, screen, inputTracker, promptTracker, cfg, &state, schedulerErrors)
		if strings.TrimSpace(state.SessionID) == "" || strings.TrimSpace(opts.promptPath) != "" {
			go injectStartupPromptAfterDelay(ctx, ptmx, prompts, artifacts, startupHeartbeatDelay, promptTracker, cfg, &state, schedulerErrors)
		}
	} else if interval.IsSet() {
		if strings.TrimSpace(state.SessionID) == "" {
			injectImmediately = true
		}
		go injectHeartbeatLoop(ctx, ptmx, prompts, artifacts, interval.Duration(), injectImmediately, cfg, &state, schedulerErrors)
	}

	outputDone := make(chan error, 1)
	go func() {
		_, err := io.Copy(io.MultiWriter(os.Stdout, runLogFile, screen), ptmx)
		if isIgnorableCopyError(err) {
			err = nil
		}
		outputDone <- err
	}()

	if term.IsTerminal(int(os.Stdin.Fd())) {
		go func() {
			_, _ = io.Copy(ptmx, trackUserInput(os.Stdin, inputTracker))
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
	case err := <-schedulerErrors:
		appendEvent(cfg.LogsDir, logEvent{
			Timestamp: time.Now().Format(time.RFC3339),
			Type:      "run_stop",
			SessionID: state.SessionID,
			Message:   err.Error(),
		})
		signalInteractiveProcess(cmd.Process, syscall.SIGINT)
		var waitErr error
		select {
		case waitErr = <-waitDone:
		case <-time.After(terminateGracePeriod):
			signalInteractiveProcess(cmd.Process, syscall.SIGKILL)
			waitErr = <-waitDone
		}
		<-outputDone
		if waitErr != nil && !isSignalExit(waitErr) {
			return waitErr
		}
		return err
	}
}

func runStatusCommand(args []string) error {
	var opts sharedOptions
	flagSet := flag.NewFlagSet("status", flag.ContinueOnError)
	flagSet.SetOutput(os.Stderr)
	registerStatusFlags(flagSet, &opts)
	flagSet.Usage = func() {
		fmt.Fprintln(flagSet.Output(), "Usage: codex-heartbeat status --workdir DIR")
		flagSet.PrintDefaults()
	}

	help, err := parseFlagSet(flagSet, args)
	if err != nil {
		return err
	}
	if help {
		return nil
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
	if err := migrateLegacyProjectDir(cfg); err != nil {
		return err
	}

	state, err := loadState(cfg.StatePath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if errors.Is(err, fs.ErrNotExist) {
		state.Workdir = cfg.Workdir
	}

	var launchSettings *launchOverrides
	program, err := loadProgramConfig(filepath.Join(cfg.Workdir, defaultProgramFilename))
	switch {
	case err == nil:
		overrides := newLaunchOverrides(program)
		if overrides.Summary() != "" {
			launchSettings = &overrides
		}
	case errors.Is(err, fs.ErrNotExist), errors.Is(err, errProgramTooLarge), errors.Is(err, errEmptyProgram):
		// Missing or unusable program metadata should not break status output.
	default:
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(statusOutput{
		workspaceState: state,
		LaunchSettings: launchSettings,
		HermesParity:   currentHermesParityStatus(),
	})
}

func currentHermesParityStatus() hermesParityStatus {
	return hermesParityStatus{
		Equivalent: false,
		Missing: []string{
			"launch-time base/developer instruction control",
			"ephemeral prefill",
			"harmless canary scoring",
		},
		TaskList: []string{
			"Add a stronger launch-time instruction channel than plain user-message reinjection.",
			"Add optional ephemeral prefill for new and resumed sessions.",
			"Add a harmless canary-scoring harness that can distinguish profile attached from profile effective.",
			"Keep the parity claim false until equivalent launch-time control and benign evaluation are both covered.",
		},
		ClaimRule: "Only claim Hermes parity after equivalent launch-time instruction control and benign evaluation coverage are both present.",
	}
}

func registerRunFlags(fs *flag.FlagSet, opts *sharedOptions) {
	fs.StringVar(&opts.workdir, "workdir", "", "Workspace directory to manage")
	fs.StringVar(&opts.promptPath, "prompt", "", "Optional heartbeat prompt file; reloaded on each emission and cached per workspace")
	fs.BoolVar(&opts.council, "council", false, "Use the council repeatedly during autoresearch instead of only as a fallback")
	fs.BoolVar(&opts.safe, "safe", false, "Do not pass --dangerously-bypass-approvals-and-sandbox to child Codex runs")
}

func registerStatusFlags(fs *flag.FlagSet, opts *sharedOptions) {
	fs.StringVar(&opts.workdir, "workdir", "", "Workspace directory to manage")
}

func parseFlagSet(fs *flag.FlagSet, args []string) (helpRequested bool, err error) {
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func prepareWorkspace(opts sharedOptions) (workspaceConfig, promptResolver, workspaceState, error) {
	if strings.TrimSpace(opts.workdir) == "" {
		return workspaceConfig{}, promptResolver{}, workspaceState{}, fmt.Errorf("--workdir is required")
	}

	cfg, err := newWorkspaceConfig(opts.workdir)
	if err != nil {
		return workspaceConfig{}, promptResolver{}, workspaceState{}, err
	}
	if err := migrateLegacyProjectDir(cfg); err != nil {
		return workspaceConfig{}, promptResolver{}, workspaceState{}, err
	}
	if err := os.MkdirAll(cfg.LogsDir, 0o755); err != nil {
		return workspaceConfig{}, promptResolver{}, workspaceState{}, fmt.Errorf("create runtime dirs: %w", err)
	}

	warning, err := ensureAutoresearchWorkspace(cfg.Workdir)
	if err != nil {
		return workspaceConfig{}, promptResolver{}, workspaceState{}, err
	}
	if warning != "" {
		fmt.Fprintln(os.Stderr, warning)
	}

	prompts, err := newPromptResolver(cfg.Workdir, opts.promptPath, cfg.ProjectDir, opts.council)
	if err != nil {
		return workspaceConfig{}, promptResolver{}, workspaceState{}, err
	}
	if err := prompts.Validate(); err != nil {
		return workspaceConfig{}, promptResolver{}, workspaceState{}, err
	}

	state, err := loadState(cfg.StatePath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return workspaceConfig{}, promptResolver{}, workspaceState{}, err
	}
	if errors.Is(err, fs.ErrNotExist) {
		state = workspaceState{Workdir: cfg.Workdir}
	}
	if state.Workdir == "" {
		state.Workdir = cfg.Workdir
	}

	return cfg, prompts, state, nil
}

func migrateLegacyProjectDir(cfg workspaceConfig) error {
	legacyDir := filepath.Join(cfg.Workdir, ".codex-heartbeat")
	legacyInfo, err := os.Stat(legacyDir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat legacy runtime dir: %w", err)
	}
	if !legacyInfo.IsDir() {
		return nil
	}

	if _, err := os.Stat(cfg.ProjectDir); err == nil {
		return nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("stat project runtime dir: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(cfg.ProjectDir), 0o755); err != nil {
		return fmt.Errorf("create projects dir: %w", err)
	}
	if err := os.Rename(legacyDir, cfg.ProjectDir); err == nil {
		return nil
	} else if !errors.Is(err, syscall.EXDEV) {
		return fmt.Errorf("migrate legacy runtime dir: %w", err)
	}

	// Cross-device moves need a copy/delete fallback instead of os.Rename.
	if err := copyTree(legacyDir, cfg.ProjectDir); err != nil {
		return fmt.Errorf("copy legacy runtime dir: %w", err)
	}
	if err := os.RemoveAll(legacyDir); err != nil {
		return fmt.Errorf("remove legacy runtime dir: %w", err)
	}
	return nil
}

func copyTree(srcRoot, dstRoot string) error {
	return filepath.WalkDir(srcRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dstRoot, relPath)

		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return os.MkdirAll(dstPath, info.Mode().Perm())
		}
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(target, dstPath)
		}
		return copyFile(path, dstPath, info.Mode().Perm())
	})
}

func copyFile(srcPath, dstPath string, perm fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return err
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return nil
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

func newPromptSource(path, projectDir string) (promptSource, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return promptSource{}, nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return promptSource{}, fmt.Errorf("resolve prompt path: %w", err)
	}

	sum := sha256.Sum256([]byte(absPath))
	return promptSource{
		path:      absPath,
		cachePath: filepath.Join(projectDir, "prompts", hex.EncodeToString(sum[:6])+".txt"),
	}, nil
}

func (p promptSource) Resolve() (string, error) {
	prompt, _, err := p.ResolveWithMetadata()
	return prompt, err
}

func (p promptSource) ResolveWithMetadata() (string, bool, error) {
	if strings.TrimSpace(p.path) == "" {
		prompt, err := normalizePromptText([]byte(defaultPrompt), "embedded heartbeat.md")
		return prompt, false, err
	}

	data, err := os.ReadFile(p.path)
	if err == nil {
		prompt, err := normalizePromptText(data, p.path)
		if err != nil {
			return "", false, err
		}
		if err := p.saveCache(prompt); err != nil {
			return "", false, err
		}
		return prompt, false, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return "", false, fmt.Errorf("read prompt file %q: %w", p.path, err)
	}

	prompt, cacheErr := p.loadCache()
	if cacheErr == nil {
		return prompt, true, nil
	}
	if errors.Is(cacheErr, fs.ErrNotExist) {
		return "", false, fmt.Errorf("prompt file %q was not found and no cached prompt is available", p.path)
	}
	return "", false, cacheErr
}

func (p promptSource) saveCache(prompt string) error {
	if strings.TrimSpace(p.cachePath) == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(p.cachePath), 0o755); err != nil {
		return fmt.Errorf("create prompt cache dir: %w", err)
	}
	if err := os.WriteFile(p.cachePath, []byte(prompt+"\n"), 0o644); err != nil {
		return fmt.Errorf("write prompt cache: %w", err)
	}
	return nil
}

func (p promptSource) loadCache() (string, error) {
	data, err := os.ReadFile(p.cachePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		return "", fmt.Errorf("read cached prompt for %q: %w", p.path, err)
	}
	prompt, err := normalizePromptText(data, p.cachePath)
	if err != nil {
		return "", fmt.Errorf("read cached prompt for %q: %w", p.path, err)
	}
	return prompt, nil
}

func normalizePromptText(data []byte, source string) (string, error) {
	prompt := strings.TrimSpace(string(data))
	if prompt == "" {
		return "", fmt.Errorf("prompt file %q is empty", source)
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

func buildInteractiveArgs(workdir, promptText, sessionID string, safe bool, sendPromptOnLaunch bool, noAltScreen bool, overrides launchOverrides) []string {
	args := []string{}
	if !safe {
		args = append(args, "--dangerously-bypass-approvals-and-sandbox")
	}
	if overrides.Profile != "" {
		args = append(args, "--profile", overrides.Profile)
	}
	if overrides.Model != "" {
		args = append(args, "--model", overrides.Model)
	}
	if overrides.ModelReasoningEffort != "" {
		args = append(args, "--config", fmt.Sprintf("model_reasoning_effort=%q", overrides.ModelReasoningEffort))
	}
	args = append(args, "--cd", workdir)
	if noAltScreen {
		args = append(args, "--no-alt-screen")
	}
	if sessionID != "" {
		args = append(args, "resume", sessionID)
		return args
	}
	if !sendPromptOnLaunch {
		return args
	}
	return append(args, promptText)
}

func newLaunchOverrides(program programConfig) launchOverrides {
	return launchOverrides{
		Profile:              strings.TrimSpace(program.Profile),
		Model:                strings.TrimSpace(program.Model),
		ModelReasoningEffort: strings.TrimSpace(program.ModelReasoningEffort),
	}
}

func (o launchOverrides) Summary() string {
	parts := []string{}
	if o.Profile != "" {
		parts = append(parts, "profile="+o.Profile)
	}
	if o.Model != "" {
		parts = append(parts, "model="+o.Model)
	}
	if o.ModelReasoningEffort != "" {
		parts = append(parts, "model_reasoning_effort="+o.ModelReasoningEffort)
	}
	return strings.Join(parts, ", ")
}

func launchSummaryOrNone(overrides launchOverrides) string {
	if summary := overrides.Summary(); summary != "" {
		return summary
	}
	return "none"
}

func interactiveLaunchBehavior(sessionID string) (sendPromptOnLaunch bool, injectHeartbeatImmediately bool) {
	hasSession := strings.TrimSpace(sessionID) != ""
	return false, hasSession
}

func resolveNoAltScreen(flagNoAltScreen, flagAltScreen bool) (bool, error) {
	if flagNoAltScreen && flagAltScreen {
		return false, fmt.Errorf("--no-alt-screen and --alt-screen cannot be used together")
	}
	if flagNoAltScreen {
		return true, nil
	}
	if flagAltScreen {
		return false, nil
	}
	return runtime.GOOS == "darwin" && strings.EqualFold(os.Getenv("TERM_PROGRAM"), "ghostty"), nil
}

func printRunBanner(cfg workspaceConfig, state workspaceState, interval, endIn durationFlag, noAltScreen bool, screenIdleHeartbeat bool) {
	mode := "new"
	if state.SessionID != "" {
		mode = "resume"
	}

	screenMode := "alt"
	if noAltScreen {
		screenMode = "inline"
	}

	var details []string
	details = append(details, fmt.Sprintf("mode=%s", mode))
	details = append(details, fmt.Sprintf("screen=%s", screenMode))
	if heartbeatMode := runHeartbeatMode(interval, screenIdleHeartbeat); heartbeatMode != "" {
		details = append(details, fmt.Sprintf("heartbeat=%s", heartbeatMode))
	}
	if endIn.IsSet() {
		details = append(details, fmt.Sprintf("end-in=%s", endIn.String()))
	}

	fmt.Fprintf(os.Stderr, "[codex-heartbeat] %s | workdir=%s\n", strings.Join(details, " | "), shortenPath(cfg.Workdir))
}

func useScreenIdleScheduler(interval durationFlag, screenIdleHeartbeat bool) bool {
	return screenIdleHeartbeat || !interval.IsSet()
}

func runHeartbeatMode(interval durationFlag, screenIdleHeartbeat bool) string {
	if useScreenIdleScheduler(interval, screenIdleHeartbeat) {
		return screenIdleHeartbeatSummary()
	}
	if interval.IsSet() {
		return interval.String()
	}
	return ""
}

func runTerminalTitle(cfg workspaceConfig) string {
	return fmt.Sprintf("codex-heartbeat | %s", shortenPath(cfg.Workdir))
}

func setTerminalTitle(title string) {
	output := terminalControlOutput()
	if output == nil {
		return
	}

	sequence := terminalTitleSequence(title)
	if sequence == "" {
		return
	}

	_, _ = io.WriteString(output, sequence)
}

func terminalControlOutput() *os.File {
	if term.IsTerminal(int(os.Stderr.Fd())) {
		return os.Stderr
	}
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return os.Stdout
	}
	return nil
}

func terminalTitleSequence(title string) string {
	sanitized := sanitizeTerminalTitle(title)
	if sanitized == "" {
		return ""
	}
	return "\033]0;" + sanitized + "\007"
}

func sanitizeTerminalTitle(title string) string {
	title = strings.Map(func(r rune) rune {
		switch {
		case r == '\a' || r == '\n' || r == '\r' || r == '\t':
			return ' '
		case r < 0x20 || r == 0x7f:
			return -1
		default:
			return r
		}
	}, title)
	return strings.Join(strings.Fields(title), " ")
}

func shortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if path == home {
		return "~"
	}
	prefix := home + string(os.PathSeparator)
	if strings.HasPrefix(path, prefix) {
		return "~" + string(os.PathSeparator) + strings.TrimPrefix(path, prefix)
	}
	return path
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

func injectStartupPromptAfterDelay(ctx context.Context, writer io.Writer, prompts promptResolver, artifacts autoresearchArtifacts, delay time.Duration, promptTracker *promptInjectionTracker, cfg workspaceConfig, state *workspaceState, errCh chan<- error) {
	if delay <= 0 {
		return
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		resolution, err := prompts.Resolve(artifacts)
		if err != nil {
			reportAsyncError(errCh, err)
			return
		}
		if err := injectPrompt(writer, resolution.Text); err == nil {
			promptTracker.Mark(time.Now())
			_ = appendExecutionNote(artifacts.ExecutionPath, fmt.Sprintf("startup heartbeat injected with prompt source `%s`", resolution.Source))
			appendEvent(cfg.LogsDir, logEvent{
				Timestamp: time.Now().Format(time.RFC3339),
				Type:      "heartbeat_injected",
				SessionID: state.SessionID,
				Message:   fmt.Sprintf("startup=%s", formatFlexibleDuration(delay)),
			})
		}
	}
}

func injectHeartbeatLoop(ctx context.Context, writer io.Writer, prompts promptResolver, artifacts autoresearchArtifacts, interval time.Duration, immediate bool, cfg workspaceConfig, state *workspaceState, errCh chan<- error) {
	if interval <= 0 {
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
			resolution, err := prompts.Resolve(artifacts)
			if err != nil {
				reportAsyncError(errCh, err)
				return
			}
			if err := injectPrompt(writer, resolution.Text); err == nil {
				_ = appendExecutionNote(artifacts.ExecutionPath, fmt.Sprintf("timed heartbeat injected with prompt source `%s`", resolution.Source))
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

func reportAsyncError(errCh chan<- error, err error) {
	if err == nil || errCh == nil {
		return
	}

	select {
	case errCh <- err:
	default:
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
	fmt.Fprintln(w, "codex-heartbeat wraps the Codex CLI and can inject heartbeat prompts based on screen state or a timer.")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  codex-heartbeat --workdir DIR [--prompt FILE] [--council] [--interval 15m] [--screen-idle-heartbeat] [--end-in 1 day]")
	fmt.Fprintln(w, "  codex-heartbeat run --workdir DIR [--prompt FILE] [--council] [--interval 15m] [--screen-idle-heartbeat] [--end-in 1 day]")
	fmt.Fprintln(w, "  codex-heartbeat status --workdir DIR")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Bare flags default to the interactive `run` command.")
	fmt.Fprintln(w, "The `status` command reports session state plus resolved `launch_settings` and `hermes_parity` details, including the safe parity `task_list`.")
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
