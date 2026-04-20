package controller

import (
	"context"
	"path/filepath"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
	"github.com/grethel-labs/kubelink-usb/internal/backup"
)

func TestBackupReconcileNotFound(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	storage := backup.NewPVCStorageWithPath(filepath.Join(t.TempDir(), "backups"))
	reconciler := &BackupReconciler{
		Client:  fake.NewClientBuilder().WithScheme(scheme).Build(),
		Scheme:  scheme,
		Storage: storage,
	}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "missing"},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("Reconcile() result = %+v, want empty", result)
	}
}

func TestBackupReconcileInitializesPhase(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	bk := &usbv1alpha1.USBBackup{
		ObjectMeta: metav1.ObjectMeta{Name: "bk-1"},
		Spec:       usbv1alpha1.USBBackupSpec{TriggerType: "manual", TriggeredBy: "admin"},
	}
	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBBackup{}).
		WithObjects(bk).
		Build()
	storage := backup.NewPVCStorageWithPath(filepath.Join(t.TempDir(), "backups"))
	reconciler := &BackupReconciler{Client: client, Scheme: scheme, Storage: storage}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: bk.Name},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Fatal("expected requeue after initialization")
	}

	var got usbv1alpha1.USBBackup
	if err := client.Get(context.Background(), types.NamespacedName{Name: bk.Name}, &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Status.Phase != "InProgress" {
		t.Fatalf("phase = %q, want %q", got.Status.Phase, "InProgress")
	}
}

func TestBackupReconcileCompletesSuccessfully(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)

	// Pre-create resources to back up.
	wl := &usbv1alpha1.USBDeviceWhitelist{
		ObjectMeta: metav1.ObjectMeta{Name: "wl-1"},
		Spec: usbv1alpha1.USBDeviceWhitelistSpec{
			Entries: []usbv1alpha1.WhitelistEntry{
				{Fingerprint: "fp-1"},
			},
		},
	}
	policy := &usbv1alpha1.USBDevicePolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "policy-1", Namespace: "default"},
	}
	approval := &usbv1alpha1.USBDeviceApproval{
		ObjectMeta: metav1.ObjectMeta{Name: "approval-1"},
		Spec:       usbv1alpha1.USBDeviceApprovalSpec{DeviceRef: usbv1alpha1.USBDeviceApprovalDeviceRef{Name: "dev-1"}, Requester: "admin"},
	}

	bk := &usbv1alpha1.USBBackup{
		ObjectMeta: metav1.ObjectMeta{Name: "bk-complete"},
		Spec:       usbv1alpha1.USBBackupSpec{TriggerType: "manual"},
		Status:     usbv1alpha1.USBBackupStatus{Phase: "InProgress"},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBBackup{}).
		WithObjects(wl, policy, approval, bk).
		Build()

	storage := backup.NewPVCStorageWithPath(filepath.Join(t.TempDir(), "backups"))
	reconciler := &BackupReconciler{Client: client, Scheme: scheme, Storage: storage}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: bk.Name},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if result.RequeueAfter != 0 {
		t.Fatalf("expected no requeue, got %v", result.RequeueAfter)
	}

	var got usbv1alpha1.USBBackup
	if err := client.Get(context.Background(), types.NamespacedName{Name: bk.Name}, &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Status.Phase != "Completed" {
		t.Fatalf("phase = %q, want %q", got.Status.Phase, "Completed")
	}
	if got.Status.Checksum == "" {
		t.Fatal("expected checksum to be set")
	}
	if got.Status.ItemCounts == nil {
		t.Fatal("expected item counts to be set")
	}
	if got.Status.ItemCounts.WhitelistEntries != 1 {
		t.Fatalf("whitelist entries = %d, want 1", got.Status.ItemCounts.WhitelistEntries)
	}
	if got.Status.ItemCounts.Policies != 1 {
		t.Fatalf("policies = %d, want 1", got.Status.ItemCounts.Policies)
	}
	if got.Status.ItemCounts.Approvals != 1 {
		t.Fatalf("approvals = %d, want 1", got.Status.ItemCounts.Approvals)
	}

	// Verify the snapshot was written to storage.
	snap, err := backup.ReadSnapshot(context.Background(), storage, bk.Name)
	if err != nil {
		t.Fatalf("ReadSnapshot() error = %v", err)
	}
	if len(snap.Data.Whitelists) != 1 {
		t.Fatalf("snapshot whitelists = %d, want 1", len(snap.Data.Whitelists))
	}
}

func TestBackupReconcileSkipsCompletedBackup(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	bk := &usbv1alpha1.USBBackup{
		ObjectMeta: metav1.ObjectMeta{Name: "bk-done"},
		Status:     usbv1alpha1.USBBackupStatus{Phase: "Completed"},
	}
	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBBackup{}).
		WithObjects(bk).
		Build()

	storage := backup.NewPVCStorageWithPath(filepath.Join(t.TempDir(), "backups"))
	reconciler := &BackupReconciler{Client: client, Scheme: scheme, Storage: storage}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: bk.Name},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("expected empty result for completed backup, got %+v", result)
	}
}

func TestBackupReconcileRetention(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	config := &usbv1alpha1.USBBackupConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "default-config"},
		Spec: usbv1alpha1.USBBackupConfigSpec{
			RetentionCount: 2,
			Destination:    usbv1alpha1.BackupDestination{Type: "pvc", PVC: &usbv1alpha1.BackupDestinationPVC{ClaimName: "test"}},
		},
	}

	// Three completed backups exist; retention is 2 so the oldest should be deleted.
	bk1 := &usbv1alpha1.USBBackup{
		ObjectMeta: metav1.ObjectMeta{Name: "bk-oldest"},
		Status:     usbv1alpha1.USBBackupStatus{Phase: "Completed"},
	}
	bk2 := &usbv1alpha1.USBBackup{
		ObjectMeta: metav1.ObjectMeta{Name: "bk-middle"},
		Status:     usbv1alpha1.USBBackupStatus{Phase: "Completed"},
	}
	bkNew := &usbv1alpha1.USBBackup{
		ObjectMeta: metav1.ObjectMeta{Name: "bk-new"},
		Spec:       usbv1alpha1.USBBackupSpec{TriggerType: "scheduled"},
		Status:     usbv1alpha1.USBBackupStatus{Phase: "InProgress"},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBBackup{}).
		WithObjects(config, bk1, bk2, bkNew).
		Build()

	storage := backup.NewPVCStorageWithPath(filepath.Join(t.TempDir(), "retention"))
	reconciler := &BackupReconciler{Client: client, Scheme: scheme, Storage: storage}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: bkNew.Name},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}

	// After reconcile of bk-new (Completed), bk-oldest should have been deleted.
	var got usbv1alpha1.USBBackup
	if err := client.Get(context.Background(), types.NamespacedName{Name: bkNew.Name}, &got); err != nil {
		t.Fatalf("Get(bk-new) error = %v", err)
	}
	if got.Status.Phase != "Completed" {
		t.Fatalf("bk-new phase = %q, want Completed", got.Status.Phase)
	}

	// Verify total backups: should have 2 completed remaining (retention=2).
	var allBackups usbv1alpha1.USBBackupList
	if err := client.List(context.Background(), &allBackups); err != nil {
		t.Fatalf("List() error = %v", err)
	}
	completedCount := 0
	for _, b := range allBackups.Items {
		if b.Status.Phase == "Completed" {
			completedCount++
		}
	}
	if completedCount > 2 {
		t.Fatalf("completed backups = %d, want <= 2", completedCount)
	}
}
