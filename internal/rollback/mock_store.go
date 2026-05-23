package rollback

import "fmt"

// MockStore is an in-memory Store used in tests.
type MockStore struct {
	data   map[string]Snapshot
	SaveFn func(snap Snapshot) error // optional override
}

// NewMockStore returns an initialised MockStore.
func NewMockStore() *MockStore {
	return &MockStore{data: make(map[string]Snapshot)}
}

func (m *MockStore) key(app, host string) string {
	return app + "@" + host
}

// Save stores the snapshot in memory, calling SaveFn first if set.
func (m *MockStore) Save(snap Snapshot) error {
	if m.SaveFn != nil {
		if err := m.SaveFn(snap); err != nil {
			return err
		}
	}
	m.data[m.key(snap.App, snap.Host)] = snap
	return nil
}

// Load retrieves a snapshot from memory.
func (m *MockStore) Load(app, host string) (*Snapshot, error) {
	snap, ok := m.data[m.key(app, host)]
	if !ok {
		return nil, fmt.Errorf("mockstore: no snapshot for %s@%s", app, host)
	}
	return &snap, nil
}
