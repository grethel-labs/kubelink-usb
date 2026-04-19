package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WhitelistEntry represents a single approved USB device fingerprint.
type WhitelistEntry struct {
	Fingerprint string       `json:"fingerprint"`
	AddedAt     *metav1.Time `json:"addedAt,omitempty"`
	AddedBy     string       `json:"addedBy,omitempty"`
	PolicyRef   string       `json:"policyRef,omitempty"`
}

// USBDeviceWhitelistSpec defines the desired state of USBDeviceWhitelist.
type USBDeviceWhitelistSpec struct {
	Entries []WhitelistEntry `json:"entries,omitempty"`
}

// USBDeviceWhitelistStatus defines the observed state of USBDeviceWhitelist.
type USBDeviceWhitelistStatus struct {
	EntryCount         int32        `json:"entryCount,omitempty"`
	LastUpdated        *metav1.Time `json:"lastUpdated,omitempty"`
	ObservedGeneration int64        `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=usbwl

// USBDeviceWhitelist is the Schema for the usbdevicewhitelists API.
type USBDeviceWhitelist struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   USBDeviceWhitelistSpec   `json:"spec,omitempty"`
	Status USBDeviceWhitelistStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// USBDeviceWhitelistList contains a list of USBDeviceWhitelist.
type USBDeviceWhitelistList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []USBDeviceWhitelist `json:"items"`
}

func init() {
	SchemeBuilder.Register(&USBDeviceWhitelist{}, &USBDeviceWhitelistList{})
}
