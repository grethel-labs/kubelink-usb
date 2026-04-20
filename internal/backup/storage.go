// Package backup implements snapshot creation, storage, and retrieval for
// USB CRD backup and restore workflows.
package backup

import (
	"context"
	"fmt"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
)

// BackupMetadata describes a stored backup without its data payload.
// Used by BackupStorage.List to enumerate available snapshots.
type BackupMetadata struct {
	Name string
	Size int64
}

// BackupStorage abstracts how backup snapshots are persisted and retrieved.
// Implementations include ConfigMap, PVC, and S3 backends selected at runtime
// via NewStorage based on USBBackupConfig destination settings.
//
// @component Storage["Backup Storage"] --> CMStorage["ConfigMapStorage"]
// @component Storage["Backup Storage"] --> PVCBackupStorage["PVCStorage"]
// @component Storage["Backup Storage"] --> S3BackupStorage["S3Storage"]
type BackupStorage interface {
	Write(ctx context.Context, name string, data []byte) error
	Read(ctx context.Context, name string) ([]byte, error)
	List(ctx context.Context) ([]BackupMetadata, error)
	Delete(ctx context.Context, name string) error
}

// NewStorage returns a BackupStorage implementation based on the destination type
// configured in the USBBackupConfig spec.
//
// Intent: Decouple callers from concrete storage backends.
// Inputs: Destination configuration from USBBackupConfigSpec.
// Outputs: BackupStorage implementation for the configured backend.
// Errors: Returns an error for unsupported or misconfigured destination types.
func NewStorage(dest usbv1alpha1.BackupDestination) (BackupStorage, error) {
	switch dest.Type {
	case "pvc":
		if dest.PVC == nil {
			return nil, fmt.Errorf("pvc destination requires pvc config")
		}
		return NewPVCStorage(dest.PVC.ClaimName, dest.PVC.SubPath), nil
	case "configmap":
		if dest.ConfigMap == nil {
			return nil, fmt.Errorf("configmap destination requires configmap config")
		}
		return NewConfigMapStorage(dest.ConfigMap.Name), nil
	case "s3":
		if dest.S3 == nil {
			return nil, fmt.Errorf("s3 destination requires s3 config")
		}
		return NewS3Storage(dest.S3.Bucket, dest.S3.Endpoint, dest.S3.Region), nil
	default:
		return nil, fmt.Errorf("unsupported backup destination type: %q", dest.Type)
	}
}
