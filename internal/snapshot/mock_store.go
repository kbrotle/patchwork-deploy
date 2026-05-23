package snapshot

import "fmt"

// MockStore is an in-memory Store for testing.
type MockStore struct {
	data   map[string]*Snapshot
	SaveFn func(appName string, snap *Snapshot) error
}

// NewMockStore returns an initialised MockStore.
func NewMockStore() *MockStore {
	return &MockStore{data: make(map[string]*Snapshot)}
}

// Save stores a snapshot in memory, delegating to SaveFn when set.
func (m *MockStore) Save(appName string, snap *Snapshot) error {
	if m.SaveFn != nil {
		return m.SaveFn(appName, snap)
	}
	m.data[appName] = snap
	return nil
}

// Load retrieves a snapshot from memory.
func (m *MockStore) Load(appName string) (*Snapshot, error) {
	snap, ok := m.data[appName]
	if !ok {
		return nil, nil
	}
	return snap, nil
}

// MustLoad returns the snapshot or panics — useful in table-driven tests.
func (m *MockStore) MustLoad(appName string) *Snapshot {
	snap, err := m.Load(appName)
	if err != nil {
		panic(fmt.Sprintf("MockStore.MustLoad: %v", err))
	}
	return snap
}
