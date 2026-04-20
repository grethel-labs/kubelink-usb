package controller

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
	kmetrics "github.com/grethel-labs/kubelink-usb/internal/metrics"
)

const usbDeviceFinalizer = "kubelink-usb.io/cleanup-export"

// USBDeviceReconciler reconciles USBDevice objects.
// It manages the initial lifecycle of discovered devices: adding cleanup
// finalizers, bootstrapping status (PendingApproval phase), and ensuring
// consistent state on the Kubernetes API.
//
// @component DeviceReconciler["USBDevice Reconciler"] --> USBDevice["USBDevice CR"]
// @flow FetchDevice["Fetch USBDevice"] --> NotFound{"NotFound?"}
// @flow NotFound -->|yes| ReturnEmpty["Return without requeue"]
// @flow NotFound -->|no| CheckDeletion{"DeletionTimestamp?"}
// @flow CheckDeletion -->|yes| RemoveFinalizer["Remove finalizer"]
// @flow CheckDeletion -->|no| EnsureFinalizer["Ensure finalizer"]
// @flow EnsureFinalizer --> InitStatus{"Phase empty?"}
// @flow InitStatus -->|yes| SetPending["Set PendingApproval"]
// @flow InitStatus -->|no| LogEvent["Log discovery"]
type USBDeviceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// Reconcile converges USBDevice state, finalizer lifecycle, and bootstrap status.
//
// Intent: Keep discovered device objects in a predictable, policy-ready state.
// Inputs: Request namespace/name and reconciliation context.
// Outputs: Empty result without requeue for current steady-state logic.
// Errors: Returns Kubernetes API or status update errors.
func (r *USBDeviceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	var device usbv1alpha1.USBDevice
	if err := r.Get(ctx, req.NamespacedName, &device); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if device.DeletionTimestamp != nil {
		if controllerutil.ContainsFinalizer(&device, usbDeviceFinalizer) {
			controllerutil.RemoveFinalizer(&device, usbDeviceFinalizer)
			if err := r.Update(ctx, &device); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(&device, usbDeviceFinalizer) {
		controllerutil.AddFinalizer(&device, usbDeviceFinalizer)
		if err := r.Update(ctx, &device); err != nil {
			return ctrl.Result{}, err
		}
	}

	if device.Status.Phase == "" {
		oldPhase := device.Status.Phase
		device.Status.Phase = "PendingApproval"
		now := metav1.Now()
		device.Status.LastSeen = &now
		device.Status.Health = "Healthy"
		if err := r.Status().Update(ctx, &device); err != nil {
			return ctrl.Result{}, err
		}
		kmetrics.UpdateDevicePhase(oldPhase, device.Status.Phase)
		kmetrics.RecordPhaseTransitionEvent(r.Recorder, &device, "device", oldPhase, device.Status.Phase)
	}

	logger.Info("discovered USB device", "device", device.Name, "node", device.Spec.NodeName, "busID", device.Spec.BusID)
	return ctrl.Result{}, nil
}

// SetupWithManager registers the USBDevice reconciler with the manager.
//
// Intent: Bind watch sources and reconciliation handler for USBDevice resources.
// Inputs: Controller-runtime manager.
// Outputs: Setup error when registration fails.
// Errors: Propagates controller builder registration errors.
func (r *USBDeviceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&usbv1alpha1.USBDevice{}).
		Complete(r)
}
