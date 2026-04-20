package backup

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
)

func TestNewStorageReturnsCorrectType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		dest    usbv1alpha1.BackupDestination
		wantErr bool
	}{
		{
			name: "pvc",
			dest: usbv1alpha1.BackupDestination{
				Type: "pvc",
				PVC:  &usbv1alpha1.BackupDestinationPVC{ClaimName: "test", SubPath: "backups/"},
			},
		},
		{
			name: "configmap",
			dest: usbv1alpha1.BackupDestination{
				Type:      "configmap",
				ConfigMap: &usbv1alpha1.BackupDestinationConfigMap{Name: "test"},
			},
		},
		{
			name: "s3",
			dest: usbv1alpha1.BackupDestination{
				Type: "s3",
				S3:   &usbv1alpha1.BackupDestinationS3{Bucket: "test", Endpoint: "localhost", Region: "us-east-1"},
			},
		},
		{
			name:    "unsupported type",
			dest:    usbv1alpha1.BackupDestination{Type: "unknown"},
			wantErr: true,
		},
		{
			name:    "pvc without config",
			dest:    usbv1alpha1.BackupDestination{Type: "pvc"},
			wantErr: true,
		},
		{
			name:    "configmap without config",
			dest:    usbv1alpha1.BackupDestination{Type: "configmap"},
			wantErr: true,
		},
		{
			name:    "s3 without config",
			dest:    usbv1alpha1.BackupDestination{Type: "s3"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s, err := NewStorage(tt.dest)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("NewStorage() error = %v", err)
			}
			if s == nil {
				t.Fatal("expected non-nil storage")
			}
		})
	}
}

func TestStorageImplementationsCRUD(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	implementations := map[string]BackupStorage{
		"configmap": NewConfigMapStorage("test"),
		"s3":        NewS3Storage("test-bucket", "localhost", "us-east-1"),
		"pvc":       NewPVCStorageWithPath(filepath.Join(tmpDir, "pvc-backups")),
	}

	for name, storage := range implementations {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			// Initially empty
			metas, err := storage.List(ctx)
			if err != nil {
				t.Fatalf("List() error = %v", err)
			}
			if len(metas) != 0 {
				t.Fatalf("List() returned %d items, want 0", len(metas))
			}

			// Write
			payload := []byte(`{"version":"v1alpha1"}`)
			if err := storage.Write(ctx, "backup-1", payload); err != nil {
				t.Fatalf("Write() error = %v", err)
			}

			// Read
			data, err := storage.Read(ctx, "backup-1")
			if err != nil {
				t.Fatalf("Read() error = %v", err)
			}
			if string(data) != string(payload) {
				t.Fatalf("Read() = %q, want %q", string(data), string(payload))
			}

			// List shows one item
			metas, err = storage.List(ctx)
			if err != nil {
				t.Fatalf("List() error = %v", err)
			}
			if len(metas) != 1 {
				t.Fatalf("List() returned %d items, want 1", len(metas))
			}
			if metas[0].Name != "backup-1" {
				t.Fatalf("List()[0].Name = %q, want %q", metas[0].Name, "backup-1")
			}

			// Write second
			if err := storage.Write(ctx, "backup-2", []byte(`{}`)); err != nil {
				t.Fatalf("Write() error = %v", err)
			}
			metas, err = storage.List(ctx)
			if err != nil {
				t.Fatalf("List() error = %v", err)
			}
			if len(metas) != 2 {
				t.Fatalf("List() returned %d items, want 2", len(metas))
			}

			// Delete
			if err := storage.Delete(ctx, "backup-1"); err != nil {
				t.Fatalf("Delete() error = %v", err)
			}
			metas, err = storage.List(ctx)
			if err != nil {
				t.Fatalf("List() error = %v", err)
			}
			if len(metas) != 1 {
				t.Fatalf("List() returned %d items, want 1", len(metas))
			}

			// Read deleted entry should fail
			_, err = storage.Read(ctx, "backup-1")
			if err == nil {
				t.Fatal("Read() expected error for deleted entry")
			}
		})
	}
}

func TestPVCStorageListEmptyDirectory(t *testing.T) {
	t.Parallel()

	s := NewPVCStorageWithPath(filepath.Join(t.TempDir(), "nonexistent"))
	metas, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(metas) != 0 {
		t.Fatalf("List() returned %d items, want 0", len(metas))
	}
}

func TestPVCStorageDeleteNonExistent(t *testing.T) {
	t.Parallel()

	s := NewPVCStorageWithPath(t.TempDir())
	if err := s.Delete(context.Background(), "nonexistent"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestPVCStorageCreatesDirectory(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "deep", "nested")
	s := NewPVCStorageWithPath(dir)
	if err := s.Write(context.Background(), "test", []byte("data")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("expected directory %s to exist: %v", dir, err)
	}
}
