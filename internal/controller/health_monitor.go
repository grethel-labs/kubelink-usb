package controller

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
)

const (
	// autoRestoreCooldown is the minimum time between consecutive auto-restores.
	autoRestoreCooldown = 10 * time.Minute
	// maxAutoRestoreRetries limits the number of failed auto-restores before giving up.
	maxAutoRestoreRetries = 3
)

// HealthCheckResult holds the outcome of a system consistency check.
type HealthCheckResult struct {
	Healthy bool
	Reason  string
}

// HealthMonitor performs periodic consistency checks and triggers auto-restore
// when the system is found to be unhealthy. It enforces a cooldown period
// (10 min) between auto-restores and a maximum retry count (3) to prevent
// restore loops.
//
// @component HealthMonitor["Health Monitor"] --> RestoreReconciler["Restore Reconciler"]
// @flow CheckHealth["Check consistency"] --> IsHealthy{"Healthy?"}
// @flow IsHealthy -->|yes| Done["No action"]
// @flow IsHealthy -->|no| CheckCooldown{"Cooldown elapsed?"}
// @flow CheckCooldown -->|no| Done
// @flow CheckCooldown -->|yes| CheckRetries{"Max retries?"}
// @flow CheckRetries -->|yes| GiveUp["Give up"]
// @flow CheckRetries -->|no| TriggerRestore["Create USBRestore CR"]
type HealthMonitor struct {
	client client.Client
}

// NewHealthMonitor creates a HealthMonitor.
func NewHealthMonitor(c client.Client) *HealthMonitor {
	return &HealthMonitor{client: c}
}

// Check evaluates system health by verifying that all consistent data resources
// are loadable and internally coherent.
//
// Intent: Detect inconsistencies in whitelists, policies, and approvals.
// Inputs: Context for Kubernetes API calls.
// Outputs: A HealthCheckResult describing whether the system is healthy.
// Errors: API-level errors are treated as unhealthy states.
func (m *HealthMonitor) Check(ctx context.Context) HealthCheckResult {
	// Verify whitelists are loadable.
	var whitelists usbv1alpha1.USBDeviceWhitelistList
	if err := m.client.List(ctx, &whitelists); err != nil {
		return HealthCheckResult{Healthy: false, Reason: fmt.Sprintf("cannot list whitelists: %v", err)}
	}

	// Verify policies are loadable.
	var policies usbv1alpha1.USBDevicePolicyList
	if err := m.client.List(ctx, &policies); err != nil {
		return HealthCheckResult{Healthy: false, Reason: fmt.Sprintf("cannot list policies: %v", err)}
	}

	// Verify approvals are loadable and reference existing policies.
	var approvals usbv1alpha1.USBDeviceApprovalList
	if err := m.client.List(ctx, &approvals); err != nil {
		return HealthCheckResult{Healthy: false, Reason: fmt.Sprintf("cannot list approvals: %v", err)}
	}

	policyIndex := make(map[string]struct{})
	for _, p := range policies.Items {
		key := p.Namespace + "/" + p.Name
		policyIndex[key] = struct{}{}
	}

	for _, a := range approvals.Items {
		if a.Spec.PolicyRef != nil {
			key := a.Spec.PolicyRef.Namespace + "/" + a.Spec.PolicyRef.Name
			if _, ok := policyIndex[key]; !ok {
				return HealthCheckResult{
					Healthy: false,
					Reason:  fmt.Sprintf("approval %q references non-existent policy %s", a.Name, key),
				}
			}
		}
	}

	return HealthCheckResult{Healthy: true, Reason: "all checks passed"}
}

// MaybeTriggerAutoRestore creates a USBRestore CR if the system is unhealthy,
// auto-restore is enabled, cooldown has elapsed, and a completed backup exists.
//
// Intent: Automate recovery without operator intervention.
// Inputs: Context and health check result.
// Outputs: True when a restore was triggered.
// Errors: Returns API errors.
func (m *HealthMonitor) MaybeTriggerAutoRestore(ctx context.Context, result HealthCheckResult) (bool, error) {
	if result.Healthy {
		return false, nil
	}

	logger := log.FromContext(ctx)

	// Check auto-restore configuration.
	var configs usbv1alpha1.USBBackupConfigList
	if err := m.client.List(ctx, &configs); err != nil || len(configs.Items) == 0 {
		return false, nil
	}
	config := &configs.Items[0]
	if !config.Spec.AutoRestore.Enabled {
		logger.Info("system unhealthy but auto-restore disabled", "reason", result.Reason)
		return false, nil
	}

	// Enforce cooldown: check if a recent restore was already created.
	var restores usbv1alpha1.USBRestoreList
	if err := m.client.List(ctx, &restores); err != nil {
		return false, err
	}

	recentAutoRestores := 0
	for _, rs := range restores.Items {
		if rs.Spec.TriggerType != "automatic" {
			continue
		}
		age := time.Since(rs.CreationTimestamp.Time)
		if age < autoRestoreCooldown {
			logger.Info("auto-restore cooldown active", "lastRestore", rs.Name, "age", age)
			return false, nil
		}
		// Only count failures within the last 24 hours for retry limiting.
		if rs.Status.Phase == "Failed" && age < 24*time.Hour {
			recentAutoRestores++
		}
	}

	if recentAutoRestores >= maxAutoRestoreRetries {
		logger.Info("max auto-restore retries reached", "retries", recentAutoRestores)
		return false, nil
	}

	// Find the most recent completed backup.
	var backups usbv1alpha1.USBBackupList
	if err := m.client.List(ctx, &backups); err != nil {
		return false, err
	}

	var latestBackup *usbv1alpha1.USBBackup
	for i := range backups.Items {
		b := &backups.Items[i]
		if b.Status.Phase != "Completed" {
			continue
		}
		if latestBackup == nil || b.CreationTimestamp.After(latestBackup.CreationTimestamp.Time) {
			latestBackup = b
		}
	}
	if latestBackup == nil {
		logger.Info("no completed backups available for auto-restore")
		return false, nil
	}

	// Create the restore CR.
	restore := &usbv1alpha1.USBRestore{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("auto-restore-%d", time.Now().Unix()),
		},
		Spec: usbv1alpha1.USBRestoreSpec{
			BackupRef:   usbv1alpha1.RestoreBackupRef{Name: latestBackup.Name},
			TriggerType: "automatic",
			TriggeredBy: "health-monitor",
		},
	}
	if err := m.client.Create(ctx, restore); err != nil {
		return false, fmt.Errorf("create auto-restore: %w", err)
	}

	// Update config status.
	config.Status.HealthStatus = "Unhealthy"
	config.Status.HealthReason = result.Reason
	if err := m.client.Status().Update(ctx, config); err != nil {
		logger.Error(err, "update backup config health status")
	}

	logger.Info("auto-restore triggered", "restore", restore.Name, "backup", latestBackup.Name, "reason", result.Reason)
	return true, nil
}
