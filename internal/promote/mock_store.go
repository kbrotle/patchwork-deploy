package promote

import "fmt"

// MockTagStore satisfies the tag.Store interface for testing.
type MockTagStore struct {
	tags map[string]map[string]string // app -> tagName -> value
	SaveFn func(app, name, value string) error
}

func NewMockTagStore() *MockTagStore {
	return &MockTagStore{tags: make(map[string]map[string]string)}
}

func (m *MockTagStore) Save(app, name, value string) error {
	if m.SaveFn != nil {
		return m.SaveFn(app, name, value)
	}
	if m.tags[app] == nil {
		m.tags[app] = make(map[string]string)
	}
	m.tags[app][name] = value
	return nil
}

func (m *MockTagStore) Load(app, name string) (string, error) {
	if v, ok := m.tags[app][name]; ok {
		return v, nil
	}
	return "", fmt.Errorf("mock tag store: %q/%q not found", app, name)
}

// MockVersionStore satisfies the version.Store interface for testing.
type MockVersionStore struct {
	versions map[string]string
	SaveFn   func(app, tag string) error
}

func NewMockVersionStore() *MockVersionStore {
	return &MockVersionStore{versions: make(map[string]string)}
}

func (m *MockVersionStore) Save(app, tag string) error {
	if m.SaveFn != nil {
		return m.SaveFn(app, tag)
	}
	m.versions[app] = tag
	return nil
}

func (m *MockVersionStore) Load(app string) (string, error) {
	if v, ok := m.versions[app]; ok {
		return v, nil
	}
	return "", fmt.Errorf("mock version store: %q not found", app)
}
