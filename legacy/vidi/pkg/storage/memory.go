package storage

import (
	"fmt"
	"sync"
)

// MemoryBackend is an in-memory storage implementation.
// Data is lost when the process stops.
type MemoryBackend struct {
	data map[string]map[string]interface{}
	mu   sync.RWMutex
}

// NewMemoryBackend creates a new in-memory backend.
func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		data: make(map[string]map[string]interface{}),
	}
}

func (m *MemoryBackend) Store(d map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	id, ok := d["id"].(string)
	if !ok || id == "" {
		return fmt.Errorf("memory backend requires an 'id' field or auto-generation logic outside")
	}
	m.data[id] = d
	return nil
}

func (m *MemoryBackend) Query(params map[string]interface{}) ([]map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []map[string]interface{}
	for _, record := range m.data {
		if matchesQuery(record, params) {
			// Return a copy to prevent external mutation
			copy := make(map[string]interface{})
			for k, v := range record {
				copy[k] = v
			}
			results = append(results, copy)
		}
	}
	return results, nil
}

func (m *MemoryBackend) Get(id string) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	record, ok := m.data[id]
	if !ok {
		return nil, fmt.Errorf("record not found: %s", id)
	}

	// Return a copy
	copy := make(map[string]interface{})
	for k, v := range record {
		copy[k] = v
	}
	return copy, nil
}

func (m *MemoryBackend) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.data[id]; !ok {
		return fmt.Errorf("record not found: %s", id)
	}
	delete(m.data, id)
	return nil
}

// matchesQuery checks if a record matches all query parameters.
func matchesQuery(record, params map[string]interface{}) bool {
	if len(params) == 0 {
		return true
	}
	for k, v := range params {
		if recVal, ok := record[k]; !ok || recVal != v {
			return false
		}
	}
	return true
}
