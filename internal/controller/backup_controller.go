package controller

import (
	"context"
	"fmt"
	"sort"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
	"github.com/grethel-labs/kubelink-usb/internal/backup"
)

// BackupReconciler reconciles USBBackup objects.
type BackupReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Storage backup.BackupStorage
}

// Reconcile drives a USBBackup through InProgress → Completed/Failed.
//
// Intent: Collect consistent configuration data and persist it through the storage backend.
// Inputs: Request namespace/name identifying the USBBackup CR.
// Outputs: Requeue after 5 s while in progress; no requeue on completion.
// Errors: Returns API or storage errors; sets phase to Failed on unrecoverable issues.
func (r *BackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var bk usbv1alpha1.USBBackup
	if err := r.Get(ctx, req.NamespacedName, &bk); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Initialize phase.
	if bk.Status.Phase == "" {
		bk.Status.Phase = "InProgress"
		if err := r.Status().Update(ctx, &bk); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	// Only process in-progress backups.
	if bk.Status.Phase != "InProgress" {
		return ctrl.Result{}, nil
	}

	// Collect resources.
	var whitelists usbv1alpha1.USBDeviceWhitelistList
	if err := r.List(ctx, &whitelists); err != nil {
		return ctrl.Result{}, fmt.Errorf("list whitelists: %w", err)
	}
	var policies usbv1alpha1.USBDevicePolicyList
	if err := r.List(ctx, &policies); err != nil {
		return ctrl.Result{}, fmt.Errorf("list policies: %w", err)
	}
	var approvals usbv1alpha1.USBDeviceApprovalList
	if err := r.List(ctx, &approvals); err != nil {
		return ctrl.Result{}, fmt.Errorf("list approvals: %w", err)
	}

	// Create and write snapshot.
	snap, err := backup.WriteSnapshot(ctx, r.Storage, bk.Name, whitelists.Items, policies.Items, approvals.Items)
	if err != nil {
		bk.Status.Phase = "Failed"
		_ = r.Status().Update(ctx, &bk)
		logger.Error(err, "backup failed")
		return ctrl.Result{}, nil
	}

	// Update status.
	now := metav1.Now()
	data, _ := backup.MarshalSnapshot(snap)
	bk.Status.Phase = "Completed"
	bk.Status.CompletedAt = &now
	bk.Status.Checksum = snap.Checksum
	bk.Status.Size = fmt.Sprintf("%dB", len(data))
	bk.Status.StorageRef = bk.Name
	bk.Status.ItemCounts = &usbv1alpha1.BackupItemCounts{
		WhitelistEntries: int32(len(whitelists.Items)),
		Policies:         int32(len(policies.Items)),
		Approvals:        int32(len(approvals.Items)),
	}
	if err := r.Status().Update(ctx, &bk); err != nil {
		return ctrl.Result{}, err
	}

	// Enforce retention.
	r.enforceRetention(ctx, logger)

	logger.Info("backup completed", "name", bk.Name, "checksum", snap.Checksum)
	return ctrl.Result{}, nil
}

// enforceRetention deletes the oldest completed backups when the total exceeds
// the retention count configured in USBBackupConfig.
func (r *BackupReconciler) enforceRetention(ctx context.Context, logger interface{ Info(string, ...interface{}) }) {
	var configs usbv1alpha1.USBBackupConfigList
	if err := r.List(ctx, &configs); err != nil || len(configs.Items) == 0 {
		return
	}
	retention := configs.Items[0].Spec.RetentionCount
	if retention <= 0 {
		return
	}

	var backups usbv1alpha1.USBBackupList
	if err := r.List(ctx, &backups); err != nil {
		return
	}

	var completed []usbv1alpha1.USBBackup
	for _, b := range backups.Items {
		if b.Status.Phase == "Completed" {
			completed = append(completed, b)
		}
	}

	if int32(len(completed)) <= retention {
		return
	}

	sort.Slice(completed, func(i, j int) bool {
		return completed[i].CreationTimestamp.Before(&completed[j].CreationTimestamp)
	})

	toDelete := int32(len(completed)) - retention
	for i := int32(0); i < toDelete; i++ {
		old := completed[i]
		if err := r.Delete(ctx, &old); err != nil && !apierrors.IsNotFound(err) {
			logger.Info("failed to delete old backup", "name", old.Name)
		}
		_ = r.Storage.Delete(ctx, old.Name)
	}
}

// SetupWithManager registers the BackupReconciler with the controller manager.
//
// Intent: Wire the backup reconcile loop into the manager lifecycle.
// Inputs: Controller-runtime manager.
// Outputs: Setup error when registration fails.
// Errors: Propagates controller builder errors.
func (r *BackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&usbv1alpha1.USBBackup{}).
		Complete(r)
}
