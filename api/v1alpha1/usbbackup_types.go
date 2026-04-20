package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackupItemCounts tracks the number of items included in a backup.
type BackupItemCounts struct {
	WhitelistEntries int32 `json:"whitelistEntries,omitempty"`
	Policies         int32 `json:"policies,omitempty"`
	Approvals        int32 `json:"approvals,omitempty"`
}

// USBBackupSpec defines the desired state of USBBackup.
type USBBackupSpec struct {
	TriggerType string `json:"triggerType"`
	TriggeredBy string `json:"triggeredBy,omitempty"`
}

// USBBackupStatus defines the observed state of USBBackup.
// Tracks backup progress, resulting checksum for integrity verification,
// and item counts for audit purposes.
//
// @state [*] --> InProgress : Backup created
// @state InProgress --> Completed : Snapshot written + checksum computed
// @state InProgress --> Failed : Storage error or collection failure
type USBBackupStatus struct {
	Phase       string            `json:"phase,omitempty"`
	CompletedAt *metav1.Time      `json:"completedAt,omitempty"`
	Size        string            `json:"size,omitempty"`
	ItemCounts  *BackupItemCounts `json:"itemCounts,omitempty"`
	StorageRef  string            `json:"storageRef,omitempty"`
	Checksum    string            `json:"checksum,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=usbbk
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Trigger",type=string,JSONPath=`.spec.triggerType`

// USBBackup is the Schema for the usbbackups API.
// Each backup captures a point-in-time snapshot of all security-relevant
// CRs (whitelists, policies, approvals) with SHA-256 integrity checksums.
// The backup storage backend is determined by the associated USBBackupConfig.
//
// @component Backup["Backup CR"] --> Storage["Backup Storage"]
// @relates USBBackupConfig ||--o{ USBBackup : "configures"
// @relates USBBackup ||--o{ USBRestore : "source for"
type USBBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   USBBackupSpec   `json:"spec,omitempty"`
	Status USBBackupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// USBBackupList contains a list of USBBackup.
type USBBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []USBBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&USBBackup{}, &USBBackupList{})
}
