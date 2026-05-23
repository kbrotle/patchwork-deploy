package status

import "fmt"

// MockStore is an in-memory Store for use in tests.
type MockStore struct {
	data   map[string]AppStatus
	SaveFn func(appName string, s AppStatus) error
}

// NewMockStore returns an initialised MockStore.
func NewMockStore() *MockStore {
	return &MockStore{data: make(map[string]AppStatus)}
}

// Save stores the status in memory, delegating to SaveFn when set.
func (m *MockStore) Save(appName string, s AppStatus) error {
	if m.SaveFn != nil {
		return m.SaveFn(appName, s)
	}
	m.data[appName] = s
	return nil
}

// Load retrieves the status from memory.
func (m *MockStore) Load(appName string) (AppStatus, error) {
	s, ok := m.data[appName]
	if !ok {
		return AppStatus{}, fmt.Errorf("mock: no record for %q", appName)
	}
	return s, nil
}
