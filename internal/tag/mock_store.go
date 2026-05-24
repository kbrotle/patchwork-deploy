package tag

import "fmt"

// MockStore is an in-memory Store for testing.
type MockStore struct {
	entries map[string]Entry
	SaveFn  func(app, tag string, entry Entry) error
}

// NewMockStore returns an initialised MockStore.
func NewMockStore() *MockStore {
	return &MockStore{entries: make(map[string]Entry)}
}

func (m *MockStore) Save(app, tag string, entry Entry) error {
	if m.SaveFn != nil {
		return m.SaveFn(app, tag, entry)
	}
	m.entries[app+":"+tag] = entry
	return nil
}

func (m *MockStore) Load(app, tag string) (Entry, error) {
	e, ok := m.entries[app+":"+tag]
	if !ok {
		return Entry{}, fmt.Errorf("tag: not found %q/%q", app, tag)
	}
	return e, nil
}

func (m *MockStore) List(app string) ([]Entry, error) {
	var out []Entry
	for k, v := range m.entries {
		if len(k) > len(app)+1 && k[:len(app)+1] == app+":" {
			out = append(out, v)
		}
	}
	return out, nil
}
