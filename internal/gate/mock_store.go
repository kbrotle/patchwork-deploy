package gate

import "fmt"

type MockStore struct {
	data   map[string]bool
	SaveFn func(app string, open bool) error
}

func NewMockStore() *MockStore {
	return &MockStore{data: make(map[string]bool)}
}

func (m *MockStore) Load(app string) (bool, error) {
	v, ok := m.data[app]
	if !ok {
		return false, fmt.Errorf("mock: no record for %q", app)
	}
	return v, nil
}

func (m *MockStore) Save(app string, open bool) error {
	if m.SaveFn != nil {
		return m.SaveFn(app, open)
	}
	m.data[app] = open
	return nil
}
