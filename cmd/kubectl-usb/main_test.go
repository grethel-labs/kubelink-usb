package main

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
)

func TestParseNamespacedName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantNS  string
		wantN   string
		wantErr bool
	}{
		{name: "valid", input: "default/conn-a", wantNS: "default", wantN: "conn-a"},
		{name: "missing slash", input: "default", wantErr: true},
		{name: "missing namespace", input: "/conn-a", wantErr: true},
		{name: "missing name", input: "default/", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseNamespacedName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Namespace != tt.wantNS || got.Name != tt.wantN {
				t.Fatalf("got %s/%s, want %s/%s", got.Namespace, got.Name, tt.wantNS, tt.wantN)
			}
		})
	}
}

func TestRunList(t *testing.T) {
	kubeClient := newFakeClient(t,
		&usbv1alpha1.USBDevice{
			ObjectMeta: metav1.ObjectMeta{Name: "dev-a"},
			Spec: usbv1alpha1.USBDeviceSpec{
				NodeName:  "node-a",
				VendorID:  "04b4",
				ProductID: "6001",
			},
			Status: usbv1alpha1.USBDeviceStatus{Phase: "Approved"},
		},
	)

	out := captureStdout(t, func() {
		if err := runList(context.Background(), kubeClient, nil); err != nil {
			t.Fatalf("runList() error = %v", err)
		}
	})

	if !strings.Contains(out, "NAME") || !strings.Contains(out, "PHASE") {
		t.Fatalf("expected header in output, got: %q", out)
	}
	if !strings.Contains(out, "dev-a") || !strings.Contains(out, "Approved") {
		t.Fatalf("expected row in output, got: %q", out)
	}
}

func TestRunApprovalAction(t *testing.T) {
	t.Setenv("USER", "test-user")
	approval := &usbv1alpha1.USBDeviceApproval{
		ObjectMeta: metav1.ObjectMeta{Name: "approval-a"},
		Spec: usbv1alpha1.USBDeviceApprovalSpec{
			Phase: "Pending",
		},
	}

	kubeClient := newFakeClient(t, approval)
	out := captureStdout(t, func() {
		if err := runApprovalAction(context.Background(), kubeClient, []string{"approval-a"}, "Approved"); err != nil {
			t.Fatalf("runApprovalAction() error = %v", err)
		}
	})
	if !strings.Contains(out, "approved approval-a") {
		t.Fatalf("unexpected output: %q", out)
	}

	var got usbv1alpha1.USBDeviceApproval
	if err := kubeClient.Get(context.Background(), types.NamespacedName{Name: "approval-a"}, &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Spec.Phase != "Approved" {
		t.Fatalf("phase=%q want Approved", got.Spec.Phase)
	}
	if got.Spec.ApprovedBy != "test-user" {
		t.Fatalf("approvedBy=%q want test-user", got.Spec.ApprovedBy)
	}
	if got.Spec.ApprovedAt == nil {
		t.Fatal("approvedAt must be set")
	}
}

func TestRunApprovalActionErrors(t *testing.T) {
	kubeClient := newFakeClient(t)

	if err := runApprovalAction(context.Background(), kubeClient, nil, "Approved"); err == nil {
		t.Fatal("expected error for missing approval name")
	}

	if err := runApprovalAction(context.Background(), kubeClient, []string{"missing"}, "Approved"); err == nil {
		t.Fatal("expected error for missing approval object")
	}
}

func TestRunConnectionAction(t *testing.T) {
	conn := &usbv1alpha1.USBConnection{
		ObjectMeta: metav1.ObjectMeta{Name: "conn-a", Namespace: "ns-a"},
		Status: usbv1alpha1.USBConnectionStatus{
			Phase: "Connected",
		},
	}
	kubeClient := newFakeClient(t, conn)

	out := captureStdout(t, func() {
		if err := runConnectionAction(context.Background(), kubeClient, []string{"ns-a/conn-a"}, "Disconnected"); err != nil {
			t.Fatalf("runConnectionAction() error = %v", err)
		}
	})
	if !strings.Contains(out, "disconnected ns-a/conn-a") {
		t.Fatalf("unexpected output: %q", out)
	}

	var got usbv1alpha1.USBConnection
	if err := kubeClient.Get(context.Background(), types.NamespacedName{Namespace: "ns-a", Name: "conn-a"}, &got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Status.Phase != "Disconnected" {
		t.Fatalf("phase=%q want Disconnected", got.Status.Phase)
	}
}

func TestRunConnectionActionErrors(t *testing.T) {
	kubeClient := newFakeClient(t)

	if err := runConnectionAction(context.Background(), kubeClient, nil, "Disconnected"); err == nil {
		t.Fatal("expected error for missing connection reference")
	}
	if err := runConnectionAction(context.Background(), kubeClient, []string{"invalid"}, "Disconnected"); err == nil {
		t.Fatal("expected parse error for invalid namespaced name")
	}
}

func TestUsage(t *testing.T) {
	out := captureStderr(t, usage)
	if !strings.Contains(out, "kubectl-usb <list|approve|deny|connect|disconnect>") {
		t.Fatalf("unexpected usage output: %q", out)
	}
}

func newFakeClient(t *testing.T, objects ...client.Object) client.Client {
	t.Helper()

	scheme := runtime.NewScheme()
	if err := usbv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme() error = %v", err)
	}
	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objects...).
		WithStatusSubresource(&usbv1alpha1.USBConnection{}).
		Build()
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = w

	done := make(chan string, 1)
	go func() {
		data, _ := io.ReadAll(r)
		done <- string(data)
	}()

	fn()

	_ = w.Close()
	os.Stdout = old
	return <-done
}

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stderr = w

	done := make(chan string, 1)
	go func() {
		data, _ := io.ReadAll(r)
		done <- string(data)
	}()

	fn()

	_ = w.Close()
	os.Stderr = old
	return <-done
}
