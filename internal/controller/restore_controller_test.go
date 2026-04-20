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

func prepareBackupForRestore(t *testing.T, tmpDir string) (backup.BackupStorage, string) {
	t.Helper()
	storage := backup.NewPVCStorageWithPath(filepath.Join(tmpDir, "restore-test"))
	ctx := context.Background()
	name := "test-backup"

	whitelists := []usbv1alpha1.USBDeviceWhitelist{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "wl-restored"},
			Spec: usbv1alpha1.USBDeviceWhitelistSpec{
				Entries: []usbv1alpha1.WhitelistEntry{
					{Fingerprint: "fp-restored"},
				},
			},
		},
	}
	policies := []usbv1alpha1.USBDevicePolicy{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "policy-restored", Namespace: "default"},
		},
	}
	approvals := []usbv1alpha1.USBDeviceApproval{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "approval-restored"},
			Spec: usbv1alpha1.USBDeviceApprovalSpec{
				DeviceRef: usbv1alpha1.USBDeviceApprovalDeviceRef{Name: "dev-1"},
				Requester: "system",
			},
		},
	}

	if _, err := backup.WriteSnapshot(ctx, storage, name, whitelists, policies, approvals); err != nil {
		t.Fatalf("WriteSnapshot() error = %v", err)
	}
	return storage, name
}

func TestRestoreReconcileNotFound(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	storage := backup.NewPVCStorageWithPath(t.TempDir())
	reconciler := &RestoreReconciler{
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
		t.Fatalf("result = %+v, want empty", result)
	}
}

func TestRestoreReconcileValidatesBackup(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	storage, backupName := prepareBackupForRestore(t, t.TempDir())

	bk := &usbv1alpha1.USBBackup{
		ObjectMeta: metav1.ObjectMeta{Name: backupName},
		Status:     usbv1alpha1.USBBackupStatus{Phase: "Completed"},
	}
	restore := &usbv1alpha1.USBRestore{
		ObjectMeta: metav1.ObjectMeta{Name: "restore-1"},
		Spec: usbv1alpha1.USBRestoreSpec{
			BackupRef:   usbv1alpha1.RestoreBackupRef{Name: backupName},
			TriggerType: "manual",
			TriggeredBy: "admin",
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBRestore{}).
		WithObjects(bk, restore).
		Build()

	reconciler := &RestoreReconciler{Client: client, Scheme: scheme, Storage: storage}

	// First reconcile: sets phase to Validating.
	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: restore.Name},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if result.RequeueAfter == 0 {
		t.Fatal("expected requeue after validation")
	}

	var got usbv1alpha1.USBRestore
	if err := client.Get(context.Background(), types.NamespacedName{Name: restore.Name}, &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Status.Phase != "Validating" {
		t.Fatalf("phase = %q, want %q", got.Status.Phase, "Validating")
	}
	if got.Status.PreRestoreHealthCheck == nil {
		t.Fatal("expected pre-restore health check to be set")
	}
}

func TestRestoreReconcileFailsForMissingBackup(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	storage := backup.NewPVCStorageWithPath(t.TempDir())

	restore := &usbv1alpha1.USBRestore{
		ObjectMeta: metav1.ObjectMeta{Name: "restore-missing"},
		Spec: usbv1alpha1.USBRestoreSpec{
			BackupRef:   usbv1alpha1.RestoreBackupRef{Name: "nonexistent"},
			TriggerType: "manual",
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBRestore{}).
		WithObjects(restore).
		Build()

	reconciler := &RestoreReconciler{Client: client, Scheme: scheme, Storage: storage}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: restore.Name},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}

	var got usbv1alpha1.USBRestore
	if err := client.Get(context.Background(), types.NamespacedName{Name: restore.Name}, &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Status.Phase != "Failed" {
		t.Fatalf("phase = %q, want %q", got.Status.Phase, "Failed")
	}
}

func TestRestoreReconcileDryRun(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	storage, backupName := prepareBackupForRestore(t, t.TempDir())

	bk := &usbv1alpha1.USBBackup{
		ObjectMeta: metav1.ObjectMeta{Name: backupName},
		Status:     usbv1alpha1.USBBackupStatus{Phase: "Completed"},
	}
	restore := &usbv1alpha1.USBRestore{
		ObjectMeta: metav1.ObjectMeta{Name: "restore-dry"},
		Spec: usbv1alpha1.USBRestoreSpec{
			BackupRef:   usbv1alpha1.RestoreBackupRef{Name: backupName},
			TriggerType: "manual",
			DryRun:      true,
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBRestore{}).
		WithObjects(bk, restore).
		Build()

	reconciler := &RestoreReconciler{Client: client, Scheme: scheme, Storage: storage}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: restore.Name},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}

	var got usbv1alpha1.USBRestore
	if err := client.Get(context.Background(), types.NamespacedName{Name: restore.Name}, &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Status.Phase != "Completed" {
		t.Fatalf("phase = %q, want %q", got.Status.Phase, "Completed")
	}
	if got.Status.CompletedAt == nil {
		t.Fatal("expected completedAt to be set")
	}
}

func TestRestoreReconcileFullLifecycle(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	storage, backupName := prepareBackupForRestore(t, t.TempDir())

	bk := &usbv1alpha1.USBBackup{
		ObjectMeta: metav1.ObjectMeta{Name: backupName},
		Status:     usbv1alpha1.USBBackupStatus{Phase: "Completed"},
	}
	restore := &usbv1alpha1.USBRestore{
		ObjectMeta: metav1.ObjectMeta{Name: "restore-full"},
		Spec: usbv1alpha1.USBRestoreSpec{
			BackupRef:   usbv1alpha1.RestoreBackupRef{Name: backupName},
			TriggerType: "manual",
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(
			&usbv1alpha1.USBRestore{},
			&usbv1alpha1.USBConnection{},
		).
		WithObjects(bk, restore).
		Build()

	reconciler := &RestoreReconciler{Client: client, Scheme: scheme, Storage: storage}
	ctx := context.Background()

	// Phase 1: "" → Validating
	_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: restore.Name}})
	if err != nil {
		t.Fatalf("phase 1 error = %v", err)
	}

	// Phase 2: Validating → Restoring
	_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: restore.Name}})
	if err != nil {
		t.Fatalf("phase 2 error = %v", err)
	}

	var afterRestore usbv1alpha1.USBRestore
	if err := client.Get(ctx, types.NamespacedName{Name: restore.Name}, &afterRestore); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if afterRestore.Status.Phase != "Restoring" {
		t.Fatalf("phase = %q, want Restoring", afterRestore.Status.Phase)
	}
	if afterRestore.Status.RestoredItems == nil {
		t.Fatal("expected restored items to be set")
	}
	if afterRestore.Status.RestoredItems.WhitelistEntries != 1 {
		t.Fatalf("restored whitelists = %d, want 1", afterRestore.Status.RestoredItems.WhitelistEntries)
	}

	// Verify the whitelists were actually created.
	var whitelists usbv1alpha1.USBDeviceWhitelistList
	if err := client.List(ctx, &whitelists); err != nil {
		t.Fatalf("List whitelists error = %v", err)
	}
	if len(whitelists.Items) != 1 {
		t.Fatalf("whitelists = %d, want 1", len(whitelists.Items))
	}

	// Phase 3: Restoring → RevalidatingConnections
	_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: restore.Name}})
	if err != nil {
		t.Fatalf("phase 3 error = %v", err)
	}

	// Phase 4: RevalidatingConnections → Completed
	_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: restore.Name}})
	if err != nil {
		t.Fatalf("phase 4 error = %v", err)
	}

	var final usbv1alpha1.USBRestore
	if err := client.Get(ctx, types.NamespacedName{Name: restore.Name}, &final); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if final.Status.Phase != "Completed" {
		t.Fatalf("phase = %q, want Completed", final.Status.Phase)
	}
	if final.Status.CompletedAt == nil {
		t.Fatal("expected completedAt to be set")
	}
}

func TestRestoreReconcileSkipsCompleted(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	storage := backup.NewPVCStorageWithPath(t.TempDir())

	restore := &usbv1alpha1.USBRestore{
		ObjectMeta: metav1.ObjectMeta{Name: "restore-done"},
		Status:     usbv1alpha1.USBRestoreStatus{Phase: "Completed"},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBRestore{}).
		WithObjects(restore).
		Build()

	reconciler := &RestoreReconciler{Client: client, Scheme: scheme, Storage: storage}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: restore.Name},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("expected empty result for completed restore, got %+v", result)
	}
}
