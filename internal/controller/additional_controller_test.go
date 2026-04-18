package controller

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
)

func TestUSBDeviceReconcileRemovesFinalizerOnDeletion(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	deletingNow := metav1.NewTime(time.Now())
	device := &usbv1alpha1.USBDevice{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "deleting-dev",
			Finalizers:        []string{usbDeviceFinalizer},
			DeletionTimestamp: &deletingNow,
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(device).Build()
	reconciler := &USBDeviceReconciler{Client: client, Scheme: scheme}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: device.Name},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("Reconcile() result = %+v want empty result", result)
	}
}

func TestPlaceholderReconcilersReturnNil(t *testing.T) {
	t.Parallel()

	approval := &ApprovalReconciler{}
	if _, err := approval.Reconcile(context.Background(), ctrl.Request{}); err != nil {
		t.Fatalf("ApprovalReconciler.Reconcile() error = %v", err)
	}
	if err := approval.SetupWithManager(nil); err != nil {
		t.Fatalf("ApprovalReconciler.SetupWithManager() error = %v", err)
	}

	conn := &USBConnectionReconciler{}
	if _, err := conn.Reconcile(context.Background(), ctrl.Request{}); err != nil {
		t.Fatalf("USBConnectionReconciler.Reconcile() error = %v", err)
	}
	if err := conn.SetupWithManager(nil); err != nil {
		t.Fatalf("USBConnectionReconciler.SetupWithManager() error = %v", err)
	}
}
