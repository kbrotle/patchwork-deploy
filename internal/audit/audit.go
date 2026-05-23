package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Event represents a single audit log entry.
type Event struct {
	Timestamp time.Time `json:"timestamp"`
	App       string    `json:"app"`
	Action    string    `json:"action"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
}

// Logger writes audit events to a JSON-lines file.
type Logger struct {
	dir string
}

// NewLogger creates an audit Logger that writes events under dir.
func NewLogger(dir string) (*Logger, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("audit: create dir: %w", err)
	}
	return &Logger{dir: dir}, nil
}

// Record appends an event to the audit log for the given app.
func (l *Logger) Record(app, action, status, message string) error {
	e := Event{
		Timestamp: time.Now().UTC(),
		App:       app,
		Action:    action,
		Status:    status,
		Message:   message,
	}

	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("audit: marshal event: %w", err)
	}

	path := filepath.Join(l.dir, app+".log")
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("audit: open log file: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s\n", data)
	return err
}

// ReadAll returns all recorded events for the given app.
func (l *Logger) ReadAll(app string) ([]Event, error) {
	path := filepath.Join(l.dir, app+".log")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("audit: read log: %w", err)
	}

	var events []Event
	for _, line := range splitLines(data) {
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("audit: parse line: %w", err)
		}
		events = append(events, e)
	}
	return events, nil
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
