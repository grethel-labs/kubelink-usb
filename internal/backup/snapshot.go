package backup

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
)

// SnapshotData holds the serializable configuration resources captured in a backup.
type SnapshotData struct {
	Whitelists []usbv1alpha1.USBDeviceWhitelist `json:"whitelists"`
	Policies   []usbv1alpha1.USBDevicePolicy    `json:"policies"`
	Approvals  []usbv1alpha1.USBDeviceApproval   `json:"approvals"`
}

// Snapshot is the top-level envelope written to backup storage.
type Snapshot struct {
	Version   string       `json:"version"`
	CreatedAt time.Time    `json:"createdAt"`
	Checksum  string       `json:"checksum"`
	Data      SnapshotData `json:"data"`
}

// CreateSnapshot builds a Snapshot from the provided resource lists.
//
// Intent: Produce a self-describing, checksummed backup payload.
// Inputs: Whitelists, policies, and approvals to include.
// Outputs: A Snapshot with computed checksum.
// Errors: Returns a JSON marshalling error if any resource is not serializable.
func CreateSnapshot(whitelists []usbv1alpha1.USBDeviceWhitelist, policies []usbv1alpha1.USBDevicePolicy, approvals []usbv1alpha1.USBDeviceApproval) (*Snapshot, error) {
	data := SnapshotData{
		Whitelists: whitelists,
		Policies:   policies,
		Approvals:  approvals,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshal snapshot data: %w", err)
	}

	checksum := fmt.Sprintf("sha256:%x", sha256.Sum256(dataBytes))

	return &Snapshot{
		Version:   "v1alpha1",
		CreatedAt: time.Now().UTC(),
		Checksum:  checksum,
		Data:      data,
	}, nil
}

// MarshalSnapshot serializes a Snapshot to JSON bytes.
//
// Intent: Produce the byte payload for storage backends.
// Inputs: A valid Snapshot.
// Outputs: JSON byte representation.
// Errors: Returns a JSON marshalling error.
func MarshalSnapshot(s *Snapshot) ([]byte, error) {
	return json.Marshal(s)
}

// UnmarshalSnapshot deserializes JSON bytes into a Snapshot.
//
// Intent: Reconstruct a Snapshot from stored data.
// Inputs: JSON bytes.
// Outputs: Parsed Snapshot struct.
// Errors: Returns a JSON unmarshalling error.
func UnmarshalSnapshot(data []byte) (*Snapshot, error) {
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot: %w", err)
	}
	return &s, nil
}

// ValidateChecksum verifies that a Snapshot's data matches its stored checksum.
//
// Intent: Detect corruption or tampering in backup data.
// Inputs: A Snapshot to validate.
// Outputs: True when the computed checksum matches the stored value.
// Errors: Returns false with no additional error detail.
func ValidateChecksum(s *Snapshot) (bool, error) {
	dataBytes, err := json.Marshal(s.Data)
	if err != nil {
		return false, fmt.Errorf("marshal snapshot data for checksum: %w", err)
	}
	expected := fmt.Sprintf("sha256:%x", sha256.Sum256(dataBytes))
	return expected == s.Checksum, nil
}

// WriteSnapshot creates a snapshot and writes it to storage.
//
// Intent: Single-step backup creation for reconciler use.
// Inputs: Context, storage backend, backup name, and resources to back up.
// Outputs: The created Snapshot.
// Errors: Returns snapshot creation or storage write errors.
func WriteSnapshot(ctx context.Context, storage BackupStorage, name string, whitelists []usbv1alpha1.USBDeviceWhitelist, policies []usbv1alpha1.USBDevicePolicy, approvals []usbv1alpha1.USBDeviceApproval) (*Snapshot, error) {
	snap, err := CreateSnapshot(whitelists, policies, approvals)
	if err != nil {
		return nil, err
	}
	data, err := MarshalSnapshot(snap)
	if err != nil {
		return nil, err
	}
	if err := storage.Write(ctx, name, data); err != nil {
		return nil, fmt.Errorf("write snapshot to storage: %w", err)
	}
	return snap, nil
}

// ReadSnapshot reads and validates a snapshot from storage.
//
// Intent: Single-step backup retrieval with integrity check.
// Inputs: Context, storage backend, and backup name.
// Outputs: The validated Snapshot.
// Errors: Returns storage read, unmarshal, or checksum validation errors.
func ReadSnapshot(ctx context.Context, storage BackupStorage, name string) (*Snapshot, error) {
	data, err := storage.Read(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("read snapshot from storage: %w", err)
	}
	snap, err := UnmarshalSnapshot(data)
	if err != nil {
		return nil, err
	}
	valid, err := ValidateChecksum(snap)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("snapshot %q checksum mismatch", name)
	}
	return snap, nil
}
