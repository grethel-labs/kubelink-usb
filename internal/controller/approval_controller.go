package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ApprovalReconciler is a phase-2 placeholder.
type ApprovalReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *ApprovalReconciler) Reconcile(context.Context, ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (r *ApprovalReconciler) SetupWithManager(mgr ctrl.Manager) error {
	_ = mgr
	return nil
}
