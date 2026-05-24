package version

import "fmt"

// MockStore is an in-memory Store for use in tests.
type MockStore struct {
	data   map[string]Entry
	SaveFn func(app string, e Entry) error
}

// NewMockStore returns an initialised MockStore.
func NewMockStore() *MockStore {
	return &MockStore{data: make(map[string]Entry)}
}

// Save stores the entry, delegating to SaveFn if set.
func (m *MockStore) Save(app string, e Entry) error {
	if m.SaveFn != nil {
		return m.SaveFn(app, e)
	}
	m.data[app] = e
	return nil
}

// Load retrieves the entry for the given app.
func (m *MockStore) Load(app string) (Entry, error) {
	e, ok := m.data[app]
	if !ok {
		return Entry{}, fmt.Errorf("version: no record for app %q", app)
	}
	return e, nil
}
