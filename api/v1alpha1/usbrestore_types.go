package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RestoreBackupRef references the backup to restore from.
type RestoreBackupRef struct {
	Name string `json:"name"`
}

// PreRestoreHealthCheck captures the health check performed before restore.
type PreRestoreHealthCheck struct {
	Status    string       `json:"status,omitempty"`
	Reason    string       `json:"reason,omitempty"`
	CheckedAt *metav1.Time `json:"checkedAt,omitempty"`
}

// ConnectionRevalidation captures results of post-restore connection validation.
type ConnectionRevalidation struct {
	Total                 int32    `json:"total,omitempty"`
	Valid                 int32    `json:"valid,omitempty"`
	Terminated            int32    `json:"terminated,omitempty"`
	TerminatedConnections []string `json:"terminatedConnections,omitempty"`
}

// RestoreItemCounts tracks the number of items restored.
type RestoreItemCounts struct {
	WhitelistEntries int32 `json:"whitelistEntries,omitempty"`
	Policies         int32 `json:"policies,omitempty"`
	Approvals        int32 `json:"approvals,omitempty"`
}

// USBRestoreSpec defines the desired state of USBRestore.
type USBRestoreSpec struct {
	BackupRef   RestoreBackupRef `json:"backupRef"`
	TriggerType string           `json:"triggerType"`
	TriggeredBy string           `json:"triggeredBy,omitempty"`
	DryRun      bool             `json:"dryRun,omitempty"`
}

// USBRestoreStatus defines the observed state of USBRestore.
// The restore follows a multi-phase lifecycle: validation, resource apply,
// connection revalidation, and completion.
//
// @state [*] --> Validating : Restore created
// @state Validating --> Restoring : Backup valid + checksum matches
// @state Validating --> Completed : DryRun mode (validation only)
// @state Validating --> Failed : Backup not found or checksum mismatch
// @state Restoring --> RevalidatingConnections : Resources applied
// @state RevalidatingConnections --> Completed : Connections validated
// @state Restoring --> Failed : Apply error
type USBRestoreStatus struct {
	Phase                  string                  `json:"phase,omitempty"`
	PreRestoreHealthCheck  *PreRestoreHealthCheck  `json:"preRestoreHealthCheck,omitempty"`
	RestoredItems          *RestoreItemCounts      `json:"restoredItems,omitempty"`
	ConnectionRevalidation *ConnectionRevalidation `json:"connectionRevalidation,omitempty"`
	CompletedAt            *metav1.Time            `json:"completedAt,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=usbrs
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Trigger",type=string,JSONPath=`.spec.triggerType`

// USBRestore is the Schema for the usbrestores API.
// A restore recreates all security-relevant CRs from a verified backup
// snapshot, revalidates active connections, and terminates any that
// reference deleted resources. Can be triggered manually or automatically
// by the HealthMonitor.
//
// @component Restore["Restore CR"] --> HealthMonitor["Health Monitor"]
type USBRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   USBRestoreSpec   `json:"spec,omitempty"`
	Status USBRestoreStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// USBRestoreList contains a list of USBRestore.
type USBRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []USBRestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&USBRestore{}, &USBRestoreList{})
}
