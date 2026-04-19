package backup

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
)

func testWhitelists() []usbv1alpha1.USBDeviceWhitelist {
	return []usbv1alpha1.USBDeviceWhitelist{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "wl-1"},
			Spec: usbv1alpha1.USBDeviceWhitelistSpec{
				Entries: []usbv1alpha1.WhitelistEntry{
					{Fingerprint: "node1-0403-6001-ABC123", AddedBy: "system"},
				},
			},
		},
	}
}

func testPolicies() []usbv1alpha1.USBDevicePolicy {
	return []usbv1alpha1.USBDevicePolicy{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "policy-1", Namespace: "default"},
			Spec: usbv1alpha1.USBDevicePolicySpec{
				Selector: usbv1alpha1.USBDevicePolicySelector{VendorID: "0403"},
			},
		},
	}
}

func testApprovals() []usbv1alpha1.USBDeviceApproval {
	return []usbv1alpha1.USBDeviceApproval{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "approval-1"},
			Spec: usbv1alpha1.USBDeviceApprovalSpec{
				DeviceRef: usbv1alpha1.USBDeviceApprovalDeviceRef{Name: "dev-1"},
				Requester: "admin",
			},
		},
	}
}

func TestCreateSnapshotAndValidateChecksum(t *testing.T) {
	t.Parallel()

	snap, err := CreateSnapshot(testWhitelists(), testPolicies(), testApprovals())
	if err != nil {
		t.Fatalf("CreateSnapshot() error = %v", err)
	}
	if snap.Version != "v1alpha1" {
		t.Fatalf("Version = %q, want %q", snap.Version, "v1alpha1")
	}
	if snap.Checksum == "" {
		t.Fatal("Checksum should not be empty")
	}
	if len(snap.Data.Whitelists) != 1 {
		t.Fatalf("Whitelists = %d, want 1", len(snap.Data.Whitelists))
	}
	if len(snap.Data.Policies) != 1 {
		t.Fatalf("Policies = %d, want 1", len(snap.Data.Policies))
	}
	if len(snap.Data.Approvals) != 1 {
		t.Fatalf("Approvals = %d, want 1", len(snap.Data.Approvals))
	}

	valid, err := ValidateChecksum(snap)
	if err != nil {
		t.Fatalf("ValidateChecksum() error = %v", err)
	}
	if !valid {
		t.Fatal("expected checksum to be valid")
	}
}

func TestValidateChecksumDetectsTampering(t *testing.T) {
	t.Parallel()

	snap, err := CreateSnapshot(testWhitelists(), testPolicies(), testApprovals())
	if err != nil {
		t.Fatalf("CreateSnapshot() error = %v", err)
	}

	snap.Data.Whitelists = nil
	valid, err := ValidateChecksum(snap)
	if err != nil {
		t.Fatalf("ValidateChecksum() error = %v", err)
	}
	if valid {
		t.Fatal("expected checksum to be invalid after tampering")
	}
}

func TestMarshalUnmarshalSnapshot(t *testing.T) {
	t.Parallel()

	snap, err := CreateSnapshot(testWhitelists(), testPolicies(), testApprovals())
	if err != nil {
		t.Fatalf("CreateSnapshot() error = %v", err)
	}

	data, err := MarshalSnapshot(snap)
	if err != nil {
		t.Fatalf("MarshalSnapshot() error = %v", err)
	}

	restored, err := UnmarshalSnapshot(data)
	if err != nil {
		t.Fatalf("UnmarshalSnapshot() error = %v", err)
	}
	if restored.Version != snap.Version {
		t.Fatalf("Version = %q, want %q", restored.Version, snap.Version)
	}
	if restored.Checksum != snap.Checksum {
		t.Fatalf("Checksum = %q, want %q", restored.Checksum, snap.Checksum)
	}
	if len(restored.Data.Whitelists) != len(snap.Data.Whitelists) {
		t.Fatalf("Whitelists = %d, want %d", len(restored.Data.Whitelists), len(snap.Data.Whitelists))
	}
}

func TestUnmarshalSnapshotInvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := UnmarshalSnapshot([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestCreateSnapshotEmptyInputs(t *testing.T) {
	t.Parallel()

	snap, err := CreateSnapshot(nil, nil, nil)
	if err != nil {
		t.Fatalf("CreateSnapshot() error = %v", err)
	}
	if snap.Data.Whitelists != nil {
		t.Fatalf("expected nil whitelists, got %v", snap.Data.Whitelists)
	}
	valid, err := ValidateChecksum(snap)
	if err != nil {
		t.Fatalf("ValidateChecksum() error = %v", err)
	}
	if !valid {
		t.Fatal("expected empty snapshot checksum to be valid")
	}
}

func TestWriteAndReadSnapshot(t *testing.T) {
	t.Parallel()

	storage := NewPVCStorageWithPath(filepath.Join(t.TempDir(), "test-snapshots"))
	ctx := context.Background()

	snap, err := WriteSnapshot(ctx, storage, "test-backup", testWhitelists(), testPolicies(), testApprovals())
	if err != nil {
		t.Fatalf("WriteSnapshot() error = %v", err)
	}
	if snap.Checksum == "" {
		t.Fatal("WriteSnapshot should produce a checksum")
	}

	restored, err := ReadSnapshot(ctx, storage, "test-backup")
	if err != nil {
		t.Fatalf("ReadSnapshot() error = %v", err)
	}
	if restored.Checksum != snap.Checksum {
		t.Fatalf("Checksum = %q, want %q", restored.Checksum, snap.Checksum)
	}
	if len(restored.Data.Whitelists) != 1 {
		t.Fatalf("Whitelists = %d, want 1", len(restored.Data.Whitelists))
	}
}

func TestReadSnapshotDetectsCorruption(t *testing.T) {
	t.Parallel()

	storage := NewPVCStorageWithPath(filepath.Join(t.TempDir(), "corrupt-test"))
	ctx := context.Background()

	snap, err := CreateSnapshot(testWhitelists(), testPolicies(), testApprovals())
	if err != nil {
		t.Fatalf("CreateSnapshot() error = %v", err)
	}
	// Tamper with the checksum before writing
	snap.Checksum = "sha256:0000000000000000000000000000000000000000000000000000000000000000"
	data, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if err := storage.Write(ctx, "corrupted", data); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	_, err = ReadSnapshot(ctx, storage, "corrupted")
	if err == nil {
		t.Fatal("expected error for corrupted snapshot")
	}
}

func TestReadSnapshotNotFound(t *testing.T) {
	t.Parallel()

	storage := NewPVCStorageWithPath(t.TempDir())
	_, err := ReadSnapshot(context.Background(), storage, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent snapshot")
	}
}
