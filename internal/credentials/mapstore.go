package credentials

import (
	"context"
	"errors"
	"sync"
)

// MapStore is an in-memory Store for tests and non-Windows dev.
// Production on Windows must use OSStore (Credential Manager).
type MapStore struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func NewMapStore() *MapStore {
	return &MapStore{data: map[string][]byte{}}
}

func clone(b []byte) []byte {
	out := make([]byte, len(b))
	copy(out, b)
	return out
}

func (m *MapStore) Set(_ context.Context, ref string, secret []byte) error {
	if err := validateRef(ref); err != nil {
		return err
	}
	if len(secret) == 0 {
		return errors.New("empty secret")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[ref] = clone(secret)
	return nil
}

func (m *MapStore) Get(_ context.Context, ref string) ([]byte, error) {
	if err := validateRef(ref); err != nil {
		return nil, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.data[ref]
	if !ok {
		return nil, ErrNotFound
	}
	return clone(v), nil
}

func (m *MapStore) Exists(_ context.Context, ref string) (bool, error) {
	if err := validateRef(ref); err != nil {
		return false, err
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[ref]
	return ok, nil
}

func (m *MapStore) Delete(_ context.Context, ref string) error {
	if err := validateRef(ref); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, ref)
	return nil
}
