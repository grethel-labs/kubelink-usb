package controller

import (
	"context"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
)

func TestHealthMonitorCheckHealthy(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	policy := &usbv1alpha1.USBDevicePolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "policy-1", Namespace: "default"},
	}
	approval := &usbv1alpha1.USBDeviceApproval{
		ObjectMeta: metav1.ObjectMeta{Name: "approval-1"},
		Spec: usbv1alpha1.USBDeviceApprovalSpec{
			DeviceRef: usbv1alpha1.USBDeviceApprovalDeviceRef{Name: "dev-1"},
			Requester: "admin",
			PolicyRef: &usbv1alpha1.USBDeviceApprovalPolicyRef{Name: "policy-1", Namespace: "default"},
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(policy, approval).
		Build()
	monitor := NewHealthMonitor(client)

	result := monitor.Check(context.Background())
	if !result.Healthy {
		t.Fatalf("Check() healthy = false, reason = %q", result.Reason)
	}
}

func TestHealthMonitorCheckUnhealthyOrphanedApproval(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	// Approval references a policy that doesn't exist.
	approval := &usbv1alpha1.USBDeviceApproval{
		ObjectMeta: metav1.ObjectMeta{Name: "orphan-approval"},
		Spec: usbv1alpha1.USBDeviceApprovalSpec{
			DeviceRef: usbv1alpha1.USBDeviceApprovalDeviceRef{Name: "dev-1"},
			Requester: "admin",
			PolicyRef: &usbv1alpha1.USBDeviceApprovalPolicyRef{Name: "missing-policy", Namespace: "default"},
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(approval).
		Build()
	monitor := NewHealthMonitor(client)

	result := monitor.Check(context.Background())
	if result.Healthy {
		t.Fatal("expected unhealthy for orphaned approval")
	}
	if result.Reason == "" {
		t.Fatal("expected reason to be set")
	}
}

func TestHealthMonitorCheckHealthyNoResources(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	monitor := NewHealthMonitor(client)

	result := monitor.Check(context.Background())
	if !result.Healthy {
		t.Fatalf("Check() healthy = false, reason = %q", result.Reason)
	}
}

func TestHealthMonitorAutoRestoreWhenUnhealthy(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	config := &usbv1alpha1.USBBackupConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "config"},
		Spec: usbv1alpha1.USBBackupConfigSpec{
			AutoRestore: usbv1alpha1.AutoRestoreConfig{Enabled: true},
			Destination: usbv1alpha1.BackupDestination{Type: "configmap", ConfigMap: &usbv1alpha1.BackupDestinationConfigMap{Name: "test"}},
		},
	}
	completedBackup := &usbv1alpha1.USBBackup{
		ObjectMeta: metav1.ObjectMeta{Name: "bk-latest"},
		Status:     usbv1alpha1.USBBackupStatus{Phase: "Completed"},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBBackupConfig{}).
		WithObjects(config, completedBackup).
		Build()
	monitor := NewHealthMonitor(client)

	unhealthy := HealthCheckResult{Healthy: false, Reason: "test unhealthy"}
	triggered, err := monitor.MaybeTriggerAutoRestore(context.Background(), unhealthy)
	if err != nil {
		t.Fatalf("MaybeTriggerAutoRestore() error = %v", err)
	}
	if !triggered {
		t.Fatal("expected auto-restore to be triggered")
	}

	// Verify a USBRestore was created.
	var restores usbv1alpha1.USBRestoreList
	if err := client.List(context.Background(), &restores); err != nil {
		t.Fatalf("List restores error = %v", err)
	}
	if len(restores.Items) != 1 {
		t.Fatalf("restores = %d, want 1", len(restores.Items))
	}
	if restores.Items[0].Spec.TriggerType != "automatic" {
		t.Fatalf("triggerType = %q, want %q", restores.Items[0].Spec.TriggerType, "automatic")
	}
}

func TestHealthMonitorNoAutoRestoreWhenDisabled(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	config := &usbv1alpha1.USBBackupConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "config"},
		Spec: usbv1alpha1.USBBackupConfigSpec{
			AutoRestore: usbv1alpha1.AutoRestoreConfig{Enabled: false},
			Destination: usbv1alpha1.BackupDestination{Type: "configmap", ConfigMap: &usbv1alpha1.BackupDestinationConfigMap{Name: "test"}},
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(config).
		Build()
	monitor := NewHealthMonitor(client)

	triggered, err := monitor.MaybeTriggerAutoRestore(context.Background(), HealthCheckResult{Healthy: false, Reason: "test"})
	if err != nil {
		t.Fatalf("MaybeTriggerAutoRestore() error = %v", err)
	}
	if triggered {
		t.Fatal("expected auto-restore NOT to be triggered when disabled")
	}
}

func TestHealthMonitorNoAutoRestoreWhenHealthy(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	monitor := NewHealthMonitor(client)

	triggered, err := monitor.MaybeTriggerAutoRestore(context.Background(), HealthCheckResult{Healthy: true})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if triggered {
		t.Fatal("expected no trigger when healthy")
	}
}

func TestHealthMonitorCooldownPreventsRestore(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	config := &usbv1alpha1.USBBackupConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "config"},
		Spec: usbv1alpha1.USBBackupConfigSpec{
			AutoRestore: usbv1alpha1.AutoRestoreConfig{Enabled: true},
			Destination: usbv1alpha1.BackupDestination{Type: "configmap", ConfigMap: &usbv1alpha1.BackupDestinationConfigMap{Name: "test"}},
		},
	}
	completedBackup := &usbv1alpha1.USBBackup{
		ObjectMeta: metav1.ObjectMeta{Name: "bk-1"},
		Status:     usbv1alpha1.USBBackupStatus{Phase: "Completed"},
	}
	// A recent automatic restore exists (within cooldown).
	recentRestore := &usbv1alpha1.USBRestore{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "recent-restore",
			CreationTimestamp: metav1.NewTime(time.Now()),
		},
		Spec: usbv1alpha1.USBRestoreSpec{
			BackupRef:   usbv1alpha1.RestoreBackupRef{Name: "bk-1"},
			TriggerType: "automatic",
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(config, completedBackup, recentRestore).
		Build()
	monitor := NewHealthMonitor(client)

	triggered, err := monitor.MaybeTriggerAutoRestore(context.Background(), HealthCheckResult{Healthy: false, Reason: "test"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if triggered {
		t.Fatal("expected cooldown to prevent auto-restore")
	}
}

func TestHealthMonitorNoBackupAvailable(t *testing.T) {
	t.Parallel()

	scheme := newTestScheme(t)
	config := &usbv1alpha1.USBBackupConfig{
		ObjectMeta: metav1.ObjectMeta{Name: "config"},
		Spec: usbv1alpha1.USBBackupConfigSpec{
			AutoRestore: usbv1alpha1.AutoRestoreConfig{Enabled: true},
			Destination: usbv1alpha1.BackupDestination{Type: "configmap", ConfigMap: &usbv1alpha1.BackupDestinationConfigMap{Name: "test"}},
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(config).
		Build()
	monitor := NewHealthMonitor(client)

	triggered, err := monitor.MaybeTriggerAutoRestore(context.Background(), HealthCheckResult{Healthy: false, Reason: "test"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if triggered {
		t.Fatal("expected no trigger when no backups available")
	}
}
