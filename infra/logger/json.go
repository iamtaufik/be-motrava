package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Log is the standard persisted log entry.
type Log struct {
	DateTime string      `json:"date_time" bson:"DateTime"`
	Message  string      `json:"message" bson:"Message"`
	Level    string      `json:"level" bson:"Level"`
	Data     interface{} `json:"data,omitempty" bson:"Data"`
	Caller   string      `json:"caller,omitempty" bson:"Caller"`
}

type handlerState struct {
	mu   sync.Mutex
	file *os.File
	logs []Log
}

type arrayJSONHandler struct {
	state  *handlerState
	attrs  []slog.Attr
	groups []string
}

// NewJSONFileLogger builds a structured logger that persists logs as a JSON array.
func NewJSONFileLogger(filePath string) (*slog.Logger, *os.File, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file: %w", err)
	}

	logs, err := loadExistingLogs(file)
	if err != nil {
		_ = file.Close()
		return nil, nil, err
	}

	handler := &arrayJSONHandler{
		state: &handlerState{
			file: file,
			logs: logs,
		},
	}

	if err := handler.persistLocked(); err != nil {
		_ = file.Close()
		return nil, nil, err
	}

	return slog.New(handler), file, nil
}

func loadExistingLogs(file *os.File) ([]Log, error) {
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("seek log file: %w", err)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("read log file: %w", err)
	}

	trimmed := strings.TrimSpace(string(content))
	if trimmed == "" {
		return []Log{}, nil
	}

	if strings.HasPrefix(trimmed, "[") {
		var logs []Log
		if err := json.Unmarshal([]byte(trimmed), &logs); err != nil {
			return nil, fmt.Errorf("decode existing json array logs: %w", err)
		}

		return logs, nil
	}

	lines := strings.Split(trimmed, "\n")
	logs := make([]Log, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry Log
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("decode existing log entry: %w", err)
		}
		logs = append(logs, entry)
	}

	return logs, nil
}

func (h *arrayJSONHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelInfo
}

func (h *arrayJSONHandler) Handle(_ context.Context, record slog.Record) error {
	entry := Log{
		DateTime: record.Time.Format(time.RFC3339Nano),
		Message:  record.Message,
		Level:    strings.ToUpper(record.Level.String()),
		Data:     h.collectData(record),
		Caller:   recordCaller(record),
	}

	h.state.mu.Lock()
	defer h.state.mu.Unlock()

	h.state.logs = append(h.state.logs, entry)
	return h.persistLocked()
}

func (h *arrayJSONHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.attrs = append(append([]slog.Attr(nil), h.attrs...), attrs...)
	return &clone
}

func (h *arrayJSONHandler) WithGroup(name string) slog.Handler {
	clone := *h
	clone.groups = append(append([]string(nil), h.groups...), name)
	return &clone
}

func (h *arrayJSONHandler) collectData(record slog.Record) interface{} {
	data := map[string]interface{}{}

	for _, attr := range h.attrs {
		addAttrToMap(data, h.groups, attr)
	}

	record.Attrs(func(attr slog.Attr) bool {
		addAttrToMap(data, h.groups, attr)
		return true
	})

	if len(data) == 0 {
		return nil
	}

	return data
}

func addAttrToMap(data map[string]interface{}, groups []string, attr slog.Attr) {
	if attr.Key == "" {
		return
	}

	key := attr.Key
	if len(groups) > 0 {
		key = strings.Join(append(groups, key), ".")
	}

	data[key] = attr.Value.Any()
}

func (h *arrayJSONHandler) persistLocked() error {
	if err := h.state.file.Truncate(0); err != nil {
		return fmt.Errorf("truncate log file: %w", err)
	}

	if _, err := h.state.file.Seek(0, 0); err != nil {
		return fmt.Errorf("seek log file: %w", err)
	}

	payload, err := json.MarshalIndent(h.state.logs, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal log array: %w", err)
	}

	if len(payload) == 0 {
		payload = []byte("[]")
	}

	if _, err := h.state.file.Write(payload); err != nil {
		return fmt.Errorf("write log file: %w", err)
	}

	if _, err := h.state.file.Write([]byte("\n")); err != nil {
		return fmt.Errorf("write log newline: %w", err)
	}

	return h.state.file.Sync()
}

func recordCaller(record slog.Record) string {
	if record.PC == 0 {
		return ""
	}

	frames := runtime.CallersFrames([]uintptr{record.PC})
	frame, _ := frames.Next()
	if frame.File == "" {
		return ""
	}

	return fmt.Sprintf("%s:%d", filepath.Base(frame.File), frame.Line)
}
