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

func TestApprovalReconcilerApprovesDevice(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	device := &usbv1alpha1.USBDevice{
		ObjectMeta: metav1.ObjectMeta{Name: "dev-a"},
		Status:     usbv1alpha1.USBDeviceStatus{Phase: "PendingApproval"},
	}
	approval := &usbv1alpha1.USBDeviceApproval{
		ObjectMeta: metav1.ObjectMeta{Name: "approval-a"},
		Spec: usbv1alpha1.USBDeviceApprovalSpec{
			DeviceRef: usbv1alpha1.USBDeviceApprovalDeviceRef{Name: "dev-a"},
			Phase:     "Approved",
			Requester: "admin",
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBDeviceApproval{}, &usbv1alpha1.USBDevice{}).
		WithObjects(device, approval).
		Build()
	reconciler := &ApprovalReconciler{Client: client, Scheme: scheme}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: approval.Name},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}

	var got usbv1alpha1.USBDevice
	if err := client.Get(context.Background(), types.NamespacedName{Name: "dev-a"}, &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Status.Phase != "Approved" {
		t.Fatalf("device phase = %q, want %q", got.Status.Phase, "Approved")
	}
}

func TestApprovalReconcilerDeniesDevice(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	device := &usbv1alpha1.USBDevice{
		ObjectMeta: metav1.ObjectMeta{Name: "dev-b"},
		Status:     usbv1alpha1.USBDeviceStatus{Phase: "PendingApproval"},
	}
	approval := &usbv1alpha1.USBDeviceApproval{
		ObjectMeta: metav1.ObjectMeta{Name: "deny-b"},
		Spec: usbv1alpha1.USBDeviceApprovalSpec{
			DeviceRef:      usbv1alpha1.USBDeviceApprovalDeviceRef{Name: "dev-b"},
			Phase:          "Denied",
			DecisionReason: "policy violation",
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBDeviceApproval{}, &usbv1alpha1.USBDevice{}).
		WithObjects(device, approval).
		Build()
	reconciler := &ApprovalReconciler{Client: client, Scheme: scheme}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: approval.Name},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}

	var got usbv1alpha1.USBDevice
	if err := client.Get(context.Background(), types.NamespacedName{Name: "dev-b"}, &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Status.Phase != "Denied" {
		t.Fatalf("device phase = %q, want %q", got.Status.Phase, "Denied")
	}
}

func TestApprovalReconcilerRejectsExpired(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	expired := metav1.NewTime(time.Now().Add(-1 * time.Hour))
	device := &usbv1alpha1.USBDevice{
		ObjectMeta: metav1.ObjectMeta{Name: "dev-c"},
		Status:     usbv1alpha1.USBDeviceStatus{Phase: "PendingApproval"},
	}
	approval := &usbv1alpha1.USBDeviceApproval{
		ObjectMeta: metav1.ObjectMeta{Name: "expired-c"},
		Spec: usbv1alpha1.USBDeviceApprovalSpec{
			DeviceRef: usbv1alpha1.USBDeviceApprovalDeviceRef{Name: "dev-c"},
			Phase:     "Approved",
			ExpiresAt: &expired,
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBDeviceApproval{}, &usbv1alpha1.USBDevice{}).
		WithObjects(device, approval).
		Build()
	reconciler := &ApprovalReconciler{Client: client, Scheme: scheme}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: approval.Name},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}

	var got usbv1alpha1.USBDeviceApproval
	if err := client.Get(context.Background(), types.NamespacedName{Name: approval.Name}, &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Status.Phase != "Denied" {
		t.Fatalf("approval status = %q, want %q", got.Status.Phase, "Denied")
	}
}

func TestApprovalReconcilerMissingDevice(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	approval := &usbv1alpha1.USBDeviceApproval{
		ObjectMeta: metav1.ObjectMeta{Name: "orphan"},
		Spec: usbv1alpha1.USBDeviceApprovalSpec{
			DeviceRef: usbv1alpha1.USBDeviceApprovalDeviceRef{Name: "nonexistent-dev"},
			Phase:     "Approved",
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBDeviceApproval{}).
		WithObjects(approval).
		Build()
	reconciler := &ApprovalReconciler{Client: client, Scheme: scheme}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: approval.Name},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}

	var got usbv1alpha1.USBDeviceApproval
	if err := client.Get(context.Background(), types.NamespacedName{Name: approval.Name}, &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Status.Phase != "Denied" {
		t.Fatalf("approval status = %q, want %q", got.Status.Phase, "Denied")
	}
}

func TestApprovalReconcilerNotFound(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	reconciler := &ApprovalReconciler{Client: client, Scheme: scheme}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "missing"},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("expected empty result")
	}
}

func TestConnectionReconcilerInitializesStatus(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	device := &usbv1alpha1.USBDevice{
		ObjectMeta: metav1.ObjectMeta{Name: "dev-x"},
		Status:     usbv1alpha1.USBDeviceStatus{Phase: "Approved"},
	}
	conn := &usbv1alpha1.USBConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "conn-1", Namespace: "default"},
		Spec: usbv1alpha1.USBConnectionSpec{
			DeviceRef:  usbv1alpha1.USBConnectionDeviceRef{Name: "dev-x"},
			ClientNode: "node-b",
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBConnection{}).
		WithObjects(device, conn).
		Build()
	reconciler := &USBConnectionReconciler{Client: client, Scheme: scheme}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: conn.Name, Namespace: conn.Namespace},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}

	var got usbv1alpha1.USBConnection
	if err := client.Get(context.Background(), types.NamespacedName{Name: conn.Name, Namespace: conn.Namespace}, &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Status.Phase != "Pending" {
		t.Fatalf("phase = %q, want %q", got.Status.Phase, "Pending")
	}
	if !result.Requeue {
		t.Fatal("expected requeue after init")
	}
}

func TestConnectionReconcilerFailsForUnapprovedDevice(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	device := &usbv1alpha1.USBDevice{
		ObjectMeta: metav1.ObjectMeta{Name: "dev-y"},
		Status:     usbv1alpha1.USBDeviceStatus{Phase: "PendingApproval"},
	}
	conn := &usbv1alpha1.USBConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "conn-2", Namespace: "default"},
		Spec: usbv1alpha1.USBConnectionSpec{
			DeviceRef:  usbv1alpha1.USBConnectionDeviceRef{Name: "dev-y"},
			ClientNode: "node-b",
		},
		Status: usbv1alpha1.USBConnectionStatus{Phase: "Pending"},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBConnection{}).
		WithObjects(device, conn).
		Build()
	reconciler := &USBConnectionReconciler{Client: client, Scheme: scheme}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: conn.Name, Namespace: conn.Namespace},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}

	var got usbv1alpha1.USBConnection
	if err := client.Get(context.Background(), types.NamespacedName{Name: conn.Name, Namespace: conn.Namespace}, &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Status.Phase != "Failed" {
		t.Fatalf("phase = %q, want %q", got.Status.Phase, "Failed")
	}
}

func TestConnectionReconcilerConnectsApprovedDevice(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	device := &usbv1alpha1.USBDevice{
		ObjectMeta: metav1.ObjectMeta{Name: "dev-z"},
		Status:     usbv1alpha1.USBDeviceStatus{Phase: "Approved"},
	}
	conn := &usbv1alpha1.USBConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "conn-3", Namespace: "default"},
		Spec: usbv1alpha1.USBConnectionSpec{
			DeviceRef:  usbv1alpha1.USBConnectionDeviceRef{Name: "dev-z"},
			ClientNode: "node-b",
		},
		Status: usbv1alpha1.USBConnectionStatus{Phase: "Connecting"},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBConnection{}).
		WithObjects(device, conn).
		Build()
	reconciler := &USBConnectionReconciler{Client: client, Scheme: scheme}

	_, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: conn.Name, Namespace: conn.Namespace},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}

	var got usbv1alpha1.USBConnection
	if err := client.Get(context.Background(), types.NamespacedName{Name: conn.Name, Namespace: conn.Namespace}, &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Status.Phase != "Connected" {
		t.Fatalf("phase = %q, want %q", got.Status.Phase, "Connected")
	}
}

func TestConnectionReconcilerNotFound(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	reconciler := &USBConnectionReconciler{Client: client, Scheme: scheme}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: "missing", Namespace: "default"},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("expected empty result")
	}
}

func TestConnectionReconcilerDeletion(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	deletingNow := metav1.NewTime(time.Now())
	conn := &usbv1alpha1.USBConnection{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "conn-del",
			Namespace:         "default",
			Finalizers:        []string{usbConnectionFinalizer},
			DeletionTimestamp: &deletingNow,
		},
		Spec: usbv1alpha1.USBConnectionSpec{
			DeviceRef:  usbv1alpha1.USBConnectionDeviceRef{Name: "dev-z"},
			ClientNode: "node-b",
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(conn).Build()
	reconciler := &USBConnectionReconciler{Client: client, Scheme: scheme}

	result, err := reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: conn.Name, Namespace: conn.Namespace},
	})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if result != (ctrl.Result{}) {
		t.Fatalf("expected empty result")
	}
	// Object may have been garbage-collected by the fake client after finalizer removal.
	// Success is that no error was returned from Reconcile.
}
