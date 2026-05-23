package audit

import "sync"

// MockLogger is an in-memory audit logger for use in tests.
type MockLogger struct {
	mu     sync.Mutex
	events []Event

	// RecordFn, if set, is called instead of the default Record behaviour.
	RecordFn func(app, action, status, message string) error
}

// NewMockLogger returns an initialised MockLogger.
func NewMockLogger() *MockLogger {
	return &MockLogger{}
}

// Record stores the event in memory (or delegates to RecordFn).
func (m *MockLogger) Record(app, action, status, message string) error {
	if m.RecordFn != nil {
		return m.RecordFn(app, action, status, message)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, Event{
		App:     app,
		Action:  action,
		Status:  status,
		Message: message,
	})
	return nil
}

// ReadAll returns all in-memory events for the given app.
func (m *MockLogger) ReadAll(app string) ([]Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Event
	for _, e := range m.events {
		if e.App == app {
			out = append(out, e)
		}
	}
	return out, nil
}

// All returns every recorded event regardless of app.
func (m *MockLogger) All() []Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	copy := make([]Event, len(m.events))
	for i, e := range m.events {
		copy[i] = e
	}
	return copy
}

// Reset clears all stored events.
func (m *MockLogger) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = nil
}
