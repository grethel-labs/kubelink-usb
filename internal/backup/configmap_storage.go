package backup

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// ConfigMapStorage is an in-memory backup storage implementation that models
// Kubernetes ConfigMap semantics. In production, this would interact with the
// Kubernetes API; the current implementation uses a map for unit-test and
// small-deployment use.
//
// @component CMStorage["ConfigMapStorage"] --> K8sAPI["Kubernetes API"]
type ConfigMapStorage struct {
	name    string
	entries map[string][]byte
	mu      sync.Mutex
}

// NewConfigMapStorage returns a ConfigMapStorage backed by the named ConfigMap.
func NewConfigMapStorage(name string) *ConfigMapStorage {
	return &ConfigMapStorage{
		name:    name,
		entries: make(map[string][]byte),
	}
}

// Write stores data under the given key.
func (s *ConfigMapStorage) Write(_ context.Context, name string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp := make([]byte, len(data))
	copy(cp, data)
	s.entries[name] = cp
	return nil
}

// Read retrieves data by key.
func (s *ConfigMapStorage) Read(_ context.Context, name string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, ok := s.entries[name]
	if !ok {
		return nil, fmt.Errorf("backup %q not found in configmap %q", name, s.name)
	}
	cp := make([]byte, len(data))
	copy(cp, data)
	return cp, nil
}

// List returns metadata for all stored entries.
func (s *ConfigMapStorage) List(_ context.Context) ([]BackupMetadata, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []BackupMetadata
	for k, v := range s.entries {
		result = append(result, BackupMetadata{Name: k, Size: int64(len(v))})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

// Delete removes a stored entry.
func (s *ConfigMapStorage) Delete(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.entries, name)
	return nil
}
