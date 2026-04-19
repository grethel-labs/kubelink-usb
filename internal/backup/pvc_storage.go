package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// PVCStorage persists backup snapshots as files in a PVC-mounted directory.
type PVCStorage struct {
	basePath string
	mu       sync.Mutex
}

// NewPVCStorage returns a PVCStorage rooted at the PVC mount path and optional sub-path.
func NewPVCStorage(claimName, subPath string) *PVCStorage {
	base := filepath.Join("/mnt", claimName)
	if subPath != "" {
		base = filepath.Join(base, subPath)
	}
	return &PVCStorage{basePath: base}
}

// NewPVCStorageWithPath returns a PVCStorage with an explicit base path.
// Useful for testing without a real PVC mount.
func NewPVCStorageWithPath(basePath string) *PVCStorage {
	return &PVCStorage{basePath: basePath}
}

func (s *PVCStorage) filePath(name string) string {
	return filepath.Join(s.basePath, name+".json")
}

// Write persists data as a JSON file under the configured base path.
func (s *PVCStorage) Write(_ context.Context, name string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.basePath, 0o750); err != nil {
		return fmt.Errorf("create backup directory: %w", err)
	}
	path := s.filePath(name)
	if err := os.WriteFile(path, data, 0o640); err != nil {
		return fmt.Errorf("write backup file %s: %w", path, err)
	}
	return nil
}

// Read loads a backup file by name.
func (s *PVCStorage) Read(_ context.Context, name string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath(name))
	if err != nil {
		return nil, fmt.Errorf("read backup file: %w", err)
	}
	return data, nil
}

// List returns metadata for all backup files in the base directory.
func (s *PVCStorage) List(_ context.Context) ([]BackupMetadata, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list backup directory: %w", err)
	}

	var result []BackupMetadata
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		name := e.Name()
		if ext := filepath.Ext(name); ext == ".json" {
			name = name[:len(name)-len(ext)]
		}
		result = append(result, BackupMetadata{Name: name, Size: info.Size()})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Name < result[j].Name })
	return result, nil
}

// Delete removes a backup file by name.
func (s *PVCStorage) Delete(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.Remove(s.filePath(name)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete backup file: %w", err)
	}
	return nil
}
