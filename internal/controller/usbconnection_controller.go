package controller

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
)

const usbConnectionFinalizer = "kubelink-usb.io/cleanup-tunnel"

// USBConnectionReconciler orchestrates USB/IP tunnel lifecycle between nodes.
type USBConnectionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile handles USBConnection tunnel lifecycle state.
//
// Intent: Orchestrate export→attach→status transitions for USB/IP tunnels.
// Inputs: Context and request identifier.
// Outputs: Empty result or requeue on transient errors.
// Errors: Returns Kubernetes API or status update errors.
func (r *USBConnectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var conn usbv1alpha1.USBConnection
	if err := r.Get(ctx, req.NamespacedName, &conn); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Handle deletion — cleanup tunnel.
	if conn.DeletionTimestamp != nil {
		if controllerutil.ContainsFinalizer(&conn, usbConnectionFinalizer) {
			logger.Info("cleaning up tunnel for deleted connection", "connection", conn.Name)
			controllerutil.RemoveFinalizer(&conn, usbConnectionFinalizer)
			if err := r.Update(ctx, &conn); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Ensure finalizer.
	if !controllerutil.ContainsFinalizer(&conn, usbConnectionFinalizer) {
		controllerutil.AddFinalizer(&conn, usbConnectionFinalizer)
		if err := r.Update(ctx, &conn); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Initialize status if empty.
	if conn.Status.Phase == "" {
		conn.Status.Phase = "Pending"
		if err := r.Status().Update(ctx, &conn); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Look up referenced device.
	var device usbv1alpha1.USBDevice
	if err := r.Get(ctx, types.NamespacedName{Name: conn.Spec.DeviceRef.Name}, &device); err != nil {
		if apierrors.IsNotFound(err) {
			return r.failConnection(ctx, &conn, "referenced device not found")
		}
		return ctrl.Result{}, err
	}

	// Only connect if device is approved.
	if device.Status.Phase != "Approved" && device.Status.Phase != "Connected" {
		return r.failConnection(ctx, &conn, fmt.Sprintf("device %q is not approved (phase: %s)", device.Name, device.Status.Phase))
	}

	switch conn.Status.Phase {
	case "Pending":
		conn.Status.Phase = "Connecting"
		if err := r.Status().Update(ctx, &conn); err != nil {
			return ctrl.Result{}, err
		}
		logger.Info("connection transitioning to Connecting", "connection", conn.Name, "device", device.Name)
		return ctrl.Result{Requeue: true}, nil

	case "Connecting":
		// Populate tunnel info from device connection info.
		if device.Status.ConnectionInfo != nil {
			conn.Status.TunnelInfo = &usbv1alpha1.USBConnectionTunnelInfo{
				ServerHost: device.Status.ConnectionInfo.Host,
				ServerPort: device.Status.ConnectionInfo.Port,
				Protocol:   "usbip",
			}
		}
		conn.Status.Phase = "Connected"
		if err := r.Status().Update(ctx, &conn); err != nil {
			return ctrl.Result{}, err
		}
		logger.Info("connection established", "connection", conn.Name)
		return ctrl.Result{}, nil

	case "Connected":
		return ctrl.Result{}, nil

	case "Failed":
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func (r *USBConnectionReconciler) failConnection(ctx context.Context, conn *usbv1alpha1.USBConnection, reason string) (ctrl.Result, error) {
	conn.Status.Phase = "Failed"
	if err := r.Status().Update(ctx, conn); err != nil {
		return ctrl.Result{}, err
	}
	logger := log.FromContext(ctx)
	logger.Info("connection failed", "connection", conn.Name, "reason", reason)
	return ctrl.Result{}, nil
}

// SetupWithManager wires USBConnectionReconciler into controller-runtime manager.
//
// Intent: Bind watch sources for USBConnection resources.
// Inputs: Controller-runtime manager.
// Outputs: Setup error when registration fails.
// Errors: Propagates controller builder registration errors.
func (r *USBConnectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&usbv1alpha1.USBConnection{}).
		Complete(r)
}
