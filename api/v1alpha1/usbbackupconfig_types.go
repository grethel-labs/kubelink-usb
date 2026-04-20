package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackupDestinationPVC configures PVC-based backup storage.
type BackupDestinationPVC struct {
	ClaimName string `json:"claimName"`
	SubPath   string `json:"subPath,omitempty"`
}

// BackupDestinationS3 configures S3-compatible backup storage.
type BackupDestinationS3 struct {
	Bucket    string                   `json:"bucket"`
	Endpoint  string                   `json:"endpoint,omitempty"`
	Region    string                   `json:"region,omitempty"`
	SecretRef *BackupDestinationSecret `json:"secretRef,omitempty"`
}

// BackupDestinationSecret references a Kubernetes secret for credentials.
type BackupDestinationSecret struct {
	Name string `json:"name"`
}

// BackupDestinationConfigMap configures ConfigMap-based backup storage.
type BackupDestinationConfigMap struct {
	Name string `json:"name"`
}

// BackupDestination defines where backups are stored.
type BackupDestination struct {
	Type      string                      `json:"type"`
	PVC       *BackupDestinationPVC       `json:"pvc,omitempty"`
	S3        *BackupDestinationS3        `json:"s3,omitempty"`
	ConfigMap *BackupDestinationConfigMap `json:"configMap,omitempty"`
}

// AutoRestoreConfig configures automatic restore behavior.
type AutoRestoreConfig struct {
	Enabled             bool            `json:"enabled,omitempty"`
	HealthCheckInterval metav1.Duration `json:"healthCheckInterval,omitempty"`
}

// USBBackupConfigSpec defines the desired backup configuration.
type USBBackupConfigSpec struct {
	Schedule       string            `json:"schedule,omitempty"`
	RetentionCount int32             `json:"retentionCount,omitempty"`
	Destination    BackupDestination `json:"destination"`
	AutoRestore    AutoRestoreConfig `json:"autoRestore,omitempty"`
}

// USBBackupConfigStatus defines the observed backup configuration state.
type USBBackupConfigStatus struct {
	LastBackupTime *metav1.Time `json:"lastBackupTime,omitempty"`
	LastBackupName string       `json:"lastBackupName,omitempty"`
	BackupCount    int32        `json:"backupCount,omitempty"`
	HealthStatus   string       `json:"healthStatus,omitempty"`
	HealthReason   string       `json:"healthReason,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=usbbc

// USBBackupConfig is the Schema for the usbbackupconfigs API.
type USBBackupConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   USBBackupConfigSpec   `json:"spec,omitempty"`
	Status USBBackupConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// USBBackupConfigList contains a list of USBBackupConfig.
type USBBackupConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []USBBackupConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&USBBackupConfig{}, &USBBackupConfigList{})
}
