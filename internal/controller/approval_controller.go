package controller

import (
	"context"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
	kmetrics "github.com/grethel-labs/kubelink-usb/internal/metrics"
	"github.com/grethel-labs/kubelink-usb/internal/security"
)

// ApprovalReconciler reconciles USBDeviceApproval objects and updates the referenced USBDevice phase.
// It processes approval decisions (approve/deny/expire) and propagates the result
// to the referenced USBDevice, driving the device through its security workflow.
//
// @component ApprovalReconciler["Approval Reconciler"] --> PolicyEngine["Policy Engine"]
// @flow FetchApproval["Fetch Approval"] --> AlreadyDone{"Already processed?"}
// @flow AlreadyDone -->|yes| SkipReturn["Return"]
// @flow AlreadyDone -->|no| CheckExpiry{"Expired?"}
// @flow CheckExpiry -->|yes| DenyExpired["Deny: expired"]
// @flow CheckExpiry -->|no| LookupDevice["Lookup USBDevice"]
// @flow LookupDevice -->|not found| DenyMissing["Deny: device missing"]
// @flow LookupDevice -->|found| PropagatePhase["Propagate to device"]
type ApprovalReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Policy   *security.Engine
	Recorder record.EventRecorder
}

// Reconcile handles USBDeviceApproval state transitions.
//
// Intent: Process approval decisions and propagate phase changes to the referenced USBDevice.
// Inputs: Context and request identifier.
// Outputs: Empty result or requeue on transient errors.
// Errors: Returns Kubernetes API or status update errors.
func (r *ApprovalReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var approval usbv1alpha1.USBDeviceApproval
	if err := r.Get(ctx, req.NamespacedName, &approval); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Skip already-processed approvals.
	if approval.Status.Phase == "Approved" || approval.Status.Phase == "Denied" {
		return ctrl.Result{}, nil
	}

	// Reject expired approvals.
	if approval.Spec.ExpiresAt != nil && approval.Spec.ExpiresAt.Time.Before(time.Now()) {
		return r.setApprovalStatus(ctx, &approval, "Denied", "system", "approval expired")
	}

	// Look up the referenced device.
	var device usbv1alpha1.USBDevice
	if err := r.Get(ctx, types.NamespacedName{Name: approval.Spec.DeviceRef.Name}, &device); err != nil {
		if apierrors.IsNotFound(err) {
			return r.setApprovalStatus(ctx, &approval, "Denied", "system", "referenced device not found")
		}
		return ctrl.Result{}, err
	}

	// Determine the phase from the spec.
	specPhase := approval.Spec.Phase
	if specPhase == "" {
		specPhase = "Approved"
	}

	// Update approval status.
	result, err := r.setApprovalStatus(ctx, &approval, specPhase, approval.Spec.ApprovedBy, approval.Spec.DecisionReason)
	if err != nil {
		return result, err
	}

	// Propagate to the device.
	var devicePhase string
	switch specPhase {
	case "Approved":
		devicePhase = "Approved"
	case "Denied":
		devicePhase = "Denied"
	default:
		return ctrl.Result{}, nil
	}

	if device.Status.Phase != devicePhase {
		oldPhase := device.Status.Phase
		device.Status.Phase = devicePhase
		if err := r.Status().Update(ctx, &device); err != nil {
			return ctrl.Result{}, err
		}
		kmetrics.UpdateDevicePhase(oldPhase, devicePhase)
		kmetrics.RecordPhaseTransitionEvent(r.Recorder, &device, "device", oldPhase, devicePhase)
		logger.Info("updated device phase from approval", "device", device.Name, "phase", devicePhase)
	}

	return ctrl.Result{}, nil
}

func (r *ApprovalReconciler) setApprovalStatus(ctx context.Context, approval *usbv1alpha1.USBDeviceApproval, phase, approvedBy, reason string) (ctrl.Result, error) {
	oldPhase := approval.Status.Phase
	now := metav1.Now()
	approval.Status.Phase = phase
	approval.Status.ApprovedBy = approvedBy
	approval.Status.ApprovedAt = &now
	approval.Status.DecisionReason = reason
	if err := r.Status().Update(ctx, approval); err != nil {
		return ctrl.Result{}, err
	}
	kmetrics.ObserveApprovalDuration(time.Since(approval.CreationTimestamp.Time))
	kmetrics.RecordPhaseTransitionEvent(r.Recorder, approval, "approval", oldPhase, phase)
	return ctrl.Result{}, nil
}

// SetupWithManager wires ApprovalReconciler into controller-runtime manager.
//
// Intent: Bind watch sources for USBDeviceApproval resources.
// Inputs: Controller-runtime manager.
// Outputs: Setup error when registration fails.
// Errors: Propagates controller builder registration errors.
func (r *ApprovalReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&usbv1alpha1.USBDeviceApproval{}).
		Complete(r)
}
