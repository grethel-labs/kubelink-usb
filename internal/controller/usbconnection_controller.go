package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// USBConnectionReconciler is a phase-2 placeholder.
type USBConnectionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *USBConnectionReconciler) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *USBConnectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	_ = mgr
	return nil
}
