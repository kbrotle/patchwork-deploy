package watch

import "fmt"

// MockStore is an in-memory Store for testing.
type MockStore struct {
	data   map[string]string
	SaveFn func(app, hash string) error
}

// NewMockStore returns an empty MockStore.
func NewMockStore() *MockStore {
	return &MockStore{data: make(map[string]string)}
}

func (m *MockStore) Load(app string) (string, error) {
	v, ok := m.data[app]
	if !ok {
		return "", nil
	}
	return v, nil
}

func (m *MockStore) Save(app, hash string) error {
	if m.SaveFn != nil {
		return m.SaveFn(app, hash)
	}
	m.data[app] = hash
	return nil
}

// Stored returns the hash saved for app, or an error if absent.
func (m *MockStore) Stored(app string) (string, error) {
	v, ok := m.data[app]
	if !ok {
		return "", fmt.Errorf("mock store: no record for %q", app)
	}
	return v, nil
}
