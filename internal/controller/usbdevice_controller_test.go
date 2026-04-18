package controller

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
)

func TestUSBDeviceReconcileNotFound(t *testing.T) {
	t.Parallel()

	reconciler := &USBDeviceReconciler{
		Client: fake.NewClientBuilder().WithScheme(newTestScheme(t)).Build(),
		Scheme: newTestScheme(t),
	}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "missing-device"},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("Reconcile() result = %+v want empty result", result)
	}
}

func TestUSBDeviceReconcileAddsFinalizerAndInitializesStatus(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	device := &usbv1alpha1.USBDevice{
		ObjectMeta: metav1.ObjectMeta{Name: "dev-a"},
		Spec: usbv1alpha1.USBDeviceSpec{
			NodeName: "node-a",
			BusID:    "1-1",
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBDevice{}).
		WithObjects(device).
		Build()
	reconciler := &USBDeviceReconciler{Client: client, Scheme: scheme}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: device.Name},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}

	var got usbv1alpha1.USBDevice
	if err := client.Get(context.Background(), types.NamespacedName{Name: device.Name}, &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !containsString(got.Finalizers, usbDeviceFinalizer) {
		t.Fatalf("finalizers = %v, expected %q to be present", got.Finalizers, usbDeviceFinalizer)
	}
	if got.Status.Phase != "PendingApproval" {
		t.Fatalf("status.phase = %q want %q", got.Status.Phase, "PendingApproval")
	}
	if got.Status.Health != "Healthy" {
		t.Fatalf("status.health = %q want %q", got.Status.Health, "Healthy")
	}
	if got.Status.LastSeen == nil {
		t.Fatal("status.lastSeen expected to be set")
	}
}

func newTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := usbv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme() error = %v", err)
	}
	return scheme
}

func containsString(items []string, expected string) bool {
	for _, item := range items {
		if item == expected {
			return true
		}
	}
	return false
}
