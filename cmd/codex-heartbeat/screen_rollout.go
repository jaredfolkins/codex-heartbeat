package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	screenRolloutTailByteLimit = 64 * 1024
	screenRolloutTailLineLimit = 96
)

type sessionRolloutInspector struct {
	mu sync.Mutex

	sessionID string
	path      string
}

type rolloutEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type rolloutPayloadHeader struct {
	Type   string `json:"type"`
	CallID string `json:"call_id,omitempty"`
}

func newSessionRolloutInspector() *sessionRolloutInspector {
	return &sessionRolloutInspector{}
}

func (i *sessionRolloutInspector) Resolve(current screenState, sessionID string) (screenState, string) {
	if current != screenStateAmbiguous || strings.TrimSpace(sessionID) == "" {
		return current, ""
	}

	path, err := i.pathForSession(sessionID)
	if err != nil || path == "" {
		return current, ""
	}

	rolloutState, reason, err := classifySessionRolloutPath(path)
	if err != nil || rolloutState == screenStateAmbiguous {
		return current, ""
	}
	return rolloutState, reason
}

func (i *sessionRolloutInspector) pathForSession(sessionID string) (string, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.sessionID != sessionID {
		i.sessionID = sessionID
		i.path = ""
	}
	if i.path != "" {
		return i.path, nil
	}

	path, err := findSessionRolloutPath(sessionID)
	if err != nil {
		return "", err
	}
	i.path = path
	return i.path, nil
}

func findSessionRolloutPath(sessionID string) (string, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return "", nil
	}

	sessionRoot := sessionRootDir()
	info, err := os.Stat(sessionRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	if !info.IsDir() {
		return "", nil
	}

	var nameMatch string
	err = filepath.WalkDir(sessionRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if nameMatch != "" || entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			return nil
		}
		if strings.Contains(entry.Name(), sessionID) {
			nameMatch = path
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if nameMatch != "" {
		return nameMatch, nil
	}

	var metaMatch string
	err = filepath.WalkDir(sessionRoot, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if metaMatch != "" || entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			return nil
		}
		record, err := readSessionMeta(path)
		if err != nil {
			return nil
		}
		if record.ID == sessionID {
			metaMatch = path
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return metaMatch, nil
}

func classifySessionRolloutPath(path string) (screenState, string, error) {
	lines, err := readJSONLTail(path, screenRolloutTailByteLimit, screenRolloutTailLineLimit)
	if err != nil {
		return screenStateAmbiguous, "", err
	}
	if len(lines) == 0 {
		return screenStateAmbiguous, "", nil
	}

	completedCalls := make(map[string]struct{})
	for idx := len(lines) - 1; idx >= 0; idx-- {
		var envelope rolloutEnvelope
		if err := json.Unmarshal([]byte(lines[idx]), &envelope); err != nil {
			continue
		}

		switch envelope.Type {
		case "compacted":
			return screenStateIdle, "rollout_compacted", nil
		case "event_msg":
			var payload rolloutPayloadHeader
			if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
				continue
			}
			switch payload.Type {
			case "task_complete", "turn_complete", "context_compacted":
				return screenStateIdle, "rollout_" + payload.Type, nil
			case "task_started", "turn_started":
				return screenStateWorking, "rollout_" + payload.Type, nil
			case "exec_command_end":
				if payload.CallID != "" {
					completedCalls[payload.CallID] = struct{}{}
				}
			}
		case "response_item":
			var payload rolloutPayloadHeader
			if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
				continue
			}
			switch payload.Type {
			case "reasoning":
				return screenStateWorking, "rollout_reasoning", nil
			case "function_call_output":
				if payload.CallID != "" {
					completedCalls[payload.CallID] = struct{}{}
				}
			case "function_call":
				if payload.CallID == "" {
					return screenStateWorking, "rollout_pending_function_call", nil
				}
				if _, ok := completedCalls[payload.CallID]; !ok {
					return screenStateWorking, "rollout_pending_function_call", nil
				}
			}
		}
	}

	return screenStateAmbiguous, "", nil
}

func readJSONLTail(path string, maxBytes, maxLines int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}

	start := int64(0)
	if maxBytes > 0 && info.Size() > int64(maxBytes) {
		start = info.Size() - int64(maxBytes)
	}
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return nil, err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if start > 0 {
		if idx := bytes.IndexByte(data, '\n'); idx >= 0 {
			data = data[idx+1:]
		} else {
			data = nil
		}
	}

	data = bytes.TrimSpace(data)
	if len(data) == 0 {
		return nil, nil
	}

	rawLines := bytes.Split(data, []byte{'\n'})
	if maxLines > 0 && len(rawLines) > maxLines {
		rawLines = rawLines[len(rawLines)-maxLines:]
	}

	lines := make([]string, 0, len(rawLines))
	for _, raw := range rawLines {
		raw = bytes.TrimSpace(raw)
		if len(raw) == 0 {
			continue
		}
		lines = append(lines, string(raw))
	}
	return lines, nil
}
