package controller

import (
	"context"
	"fmt"
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

// RestoreReconciler reconciles USBRestore objects.
type RestoreReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Storage backup.BackupStorage
}

// Reconcile drives a USBRestore through Validating → Restoring → RevalidatingConnections → Completed/Failed.
//
// Intent: Apply a backup snapshot and revalidate all active connections.
// Inputs: Request namespace/name identifying the USBRestore CR.
// Outputs: Requeue while processing; no requeue on completion.
// Errors: Returns API or storage errors; sets phase to Failed on unrecoverable issues.
func (r *RestoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var restore usbv1alpha1.USBRestore
	if err := r.Get(ctx, req.NamespacedName, &restore); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	switch restore.Status.Phase {
	case "":
		return r.phaseValidating(ctx, &restore)
	case "Validating":
		return r.phaseRestoring(ctx, &restore)
	case "Restoring":
		return r.phaseRevalidating(ctx, &restore)
	case "RevalidatingConnections":
		return r.phaseComplete(ctx, &restore)
	case "Completed", "Failed":
		return ctrl.Result{}, nil
	default:
		logger.Info("unknown restore phase", "phase", restore.Status.Phase)
		return ctrl.Result{}, nil
	}
}

func (r *RestoreReconciler) phaseValidating(ctx context.Context, restore *usbv1alpha1.USBRestore) (ctrl.Result, error) {
	restore.Status.Phase = "Validating"
	now := metav1.Now()
	restore.Status.PreRestoreHealthCheck = &usbv1alpha1.PreRestoreHealthCheck{
		Status:    "Checking",
		CheckedAt: &now,
	}
	if err := r.Status().Update(ctx, restore); err != nil {
		return ctrl.Result{}, err
	}

	// Verify referenced backup exists and is completed.
	var bk usbv1alpha1.USBBackup
	if err := r.Get(ctx, client.ObjectKey{Name: restore.Spec.BackupRef.Name}, &bk); err != nil {
		return r.failRestore(ctx, restore, fmt.Sprintf("backup %q not found: %v", restore.Spec.BackupRef.Name, err))
	}
	if bk.Status.Phase != "Completed" {
		return r.failRestore(ctx, restore, fmt.Sprintf("backup %q is not completed (phase: %s)", bk.Name, bk.Status.Phase))
	}

	// Verify snapshot can be read and checksums match.
	_, err := backup.ReadSnapshot(ctx, r.Storage, bk.Name)
	if err != nil {
		return r.failRestore(ctx, restore, fmt.Sprintf("snapshot validation failed: %v", err))
	}

	// DryRun mode: report what would happen and complete.
	if restore.Spec.DryRun {
		restore.Status.Phase = "Completed"
		restore.Status.PreRestoreHealthCheck.Status = "DryRun"
		restore.Status.PreRestoreHealthCheck.Reason = "dry-run validation passed"
		completedAt := metav1.Now()
		restore.Status.CompletedAt = &completedAt
		if err := r.Status().Update(ctx, restore); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	restore.Status.PreRestoreHealthCheck.Status = "Validated"
	restore.Status.PreRestoreHealthCheck.Reason = "backup exists and checksum valid"
	if err := r.Status().Update(ctx, restore); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: time.Second}, nil
}

func (r *RestoreReconciler) phaseRestoring(ctx context.Context, restore *usbv1alpha1.USBRestore) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	restore.Status.Phase = "Restoring"
	if err := r.Status().Update(ctx, restore); err != nil {
		return ctrl.Result{}, err
	}

	snap, err := backup.ReadSnapshot(ctx, r.Storage, restore.Spec.BackupRef.Name)
	if err != nil {
		return r.failRestore(ctx, restore, fmt.Sprintf("read snapshot: %v", err))
	}

	// Delete existing whitelists and recreate from backup.
	var existingWLs usbv1alpha1.USBDeviceWhitelistList
	if err := r.List(ctx, &existingWLs); err != nil {
		return ctrl.Result{}, err
	}
	for i := range existingWLs.Items {
		if err := r.Delete(ctx, &existingWLs.Items[i]); err != nil && !apierrors.IsNotFound(err) {
			logger.Error(err, "delete whitelist", "name", existingWLs.Items[i].Name)
		}
	}
	for i := range snap.Data.Whitelists {
		wl := snap.Data.Whitelists[i].DeepCopy()
		wl.ResourceVersion = ""
		wl.UID = ""
		if err := r.Create(ctx, wl); err != nil {
			logger.Error(err, "create whitelist", "name", wl.Name)
		}
	}

	// Delete existing policies and recreate from backup.
	var existingPolicies usbv1alpha1.USBDevicePolicyList
	if err := r.List(ctx, &existingPolicies); err != nil {
		return ctrl.Result{}, err
	}
	for i := range existingPolicies.Items {
		if err := r.Delete(ctx, &existingPolicies.Items[i]); err != nil && !apierrors.IsNotFound(err) {
			logger.Error(err, "delete policy", "name", existingPolicies.Items[i].Name)
		}
	}
	for i := range snap.Data.Policies {
		pol := snap.Data.Policies[i].DeepCopy()
		pol.ResourceVersion = ""
		pol.UID = ""
		if err := r.Create(ctx, pol); err != nil {
			logger.Error(err, "create policy", "name", pol.Name)
		}
	}

	// Delete existing approvals and recreate from backup.
	var existingApprovals usbv1alpha1.USBDeviceApprovalList
	if err := r.List(ctx, &existingApprovals); err != nil {
		return ctrl.Result{}, err
	}
	for i := range existingApprovals.Items {
		if err := r.Delete(ctx, &existingApprovals.Items[i]); err != nil && !apierrors.IsNotFound(err) {
			logger.Error(err, "delete approval", "name", existingApprovals.Items[i].Name)
		}
	}
	for i := range snap.Data.Approvals {
		appr := snap.Data.Approvals[i].DeepCopy()
		appr.ResourceVersion = ""
		appr.UID = ""
		if err := r.Create(ctx, appr); err != nil {
			logger.Error(err, "create approval", "name", appr.Name)
		}
	}

	restore.Status.RestoredItems = &usbv1alpha1.RestoreItemCounts{
		WhitelistEntries: int32(len(snap.Data.Whitelists)),
		Policies:         int32(len(snap.Data.Policies)),
		Approvals:        int32(len(snap.Data.Approvals)),
	}
	if err := r.Status().Update(ctx, restore); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: time.Second}, nil
}

func (r *RestoreReconciler) phaseRevalidating(ctx context.Context, restore *usbv1alpha1.USBRestore) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	restore.Status.Phase = "RevalidatingConnections"
	if err := r.Status().Update(ctx, restore); err != nil {
		return ctrl.Result{}, err
	}

	// List all connections and validate each.
	var connections usbv1alpha1.USBConnectionList
	if err := r.List(ctx, &connections); err != nil {
		return ctrl.Result{}, err
	}

	// Load current whitelists for validation.
	var whitelists usbv1alpha1.USBDeviceWhitelistList
	if err := r.List(ctx, &whitelists); err != nil {
		return ctrl.Result{}, err
	}
	fingerprints := make(map[string]struct{})
	for _, wl := range whitelists.Items {
		for _, e := range wl.Spec.Entries {
			fingerprints[e.Fingerprint] = struct{}{}
		}
	}

	revalidation := &usbv1alpha1.ConnectionRevalidation{
		Total: int32(len(connections.Items)),
	}

	for i := range connections.Items {
		conn := &connections.Items[i]
		// Check if referenced device exists.
		var device usbv1alpha1.USBDevice
		err := r.Get(ctx, client.ObjectKey{Name: conn.Spec.DeviceRef.Name}, &device)
		if err != nil {
			// Device not found → terminate connection.
			conn.Status.Phase = "Failed"
			if statusErr := r.Status().Update(ctx, conn); statusErr != nil {
				logger.Error(statusErr, "terminate connection status update", "name", conn.Name)
			}
			revalidation.Terminated++
			revalidation.TerminatedConnections = append(revalidation.TerminatedConnections, conn.Name)
			continue
		}
		revalidation.Valid++
	}

	restore.Status.ConnectionRevalidation = revalidation
	if err := r.Status().Update(ctx, restore); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{RequeueAfter: time.Second}, nil
}

func (r *RestoreReconciler) phaseComplete(ctx context.Context, restore *usbv1alpha1.USBRestore) (ctrl.Result, error) {
	now := metav1.Now()
	restore.Status.Phase = "Completed"
	restore.Status.CompletedAt = &now
	if err := r.Status().Update(ctx, restore); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *RestoreReconciler) failRestore(ctx context.Context, restore *usbv1alpha1.USBRestore, reason string) (ctrl.Result, error) {
	restore.Status.Phase = "Failed"
	if restore.Status.PreRestoreHealthCheck != nil {
		restore.Status.PreRestoreHealthCheck.Status = "Failed"
		restore.Status.PreRestoreHealthCheck.Reason = reason
	}
	if err := r.Status().Update(ctx, restore); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager registers the RestoreReconciler with the controller manager.
//
// Intent: Wire the restore reconcile loop into the manager lifecycle.
// Inputs: Controller-runtime manager.
// Outputs: Setup error when registration fails.
// Errors: Propagates controller builder errors.
func (r *RestoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&usbv1alpha1.USBRestore{}).
		Complete(r)
}
