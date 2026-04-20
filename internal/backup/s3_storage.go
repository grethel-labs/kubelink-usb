package backup

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// S3Storage is a placeholder implementation for S3-compatible backup storage.
// A production implementation would use an S3 SDK; this stub stores data in
// memory so the interface contract is exercised in tests.
//
// @component S3BackupStorage["S3Storage"] --> S3Bucket["S3 Endpoint"]
type S3Storage struct {
	bucket   string
	endpoint string
	region   string
	entries  map[string][]byte
	mu       sync.Mutex
}

// NewS3Storage returns an S3Storage stub for the given bucket configuration.
func NewS3Storage(bucket, endpoint, region string) *S3Storage {
	return &S3Storage{
		bucket:   bucket,
		endpoint: endpoint,
		region:   region,
		entries:  make(map[string][]byte),
	}
}

// Write stores data under the given key.
func (s *S3Storage) Write(_ context.Context, name string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cp := make([]byte, len(data))
	copy(cp, data)
	s.entries[name] = cp
	return nil
}

// Read retrieves data by key.
func (s *S3Storage) Read(_ context.Context, name string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, ok := s.entries[name]
	if !ok {
		return nil, fmt.Errorf("backup %q not found in s3 bucket %q", name, s.bucket)
	}
	cp := make([]byte, len(data))
	copy(cp, data)
	return cp, nil
}

// List returns metadata for all stored entries.
func (s *S3Storage) List(_ context.Context) ([]BackupMetadata, error) {
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
func (s *S3Storage) Delete(_ context.Context, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.entries, name)
	return nil
}
