package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// USBDeviceSpec defines the desired state of USBDevice.
type USBDeviceSpec struct {
	BusID        string `json:"busID"`
	DevicePath   string `json:"devicePath"`
	NodeName     string `json:"nodeName"`
	VendorID     string `json:"vendorID"`
	ProductID    string `json:"productID"`
	SerialNumber string `json:"serialNumber,omitempty"`
	DeviceClass  string `json:"deviceClass,omitempty"`
	Description  string `json:"description,omitempty"`
}

// USBDeviceConnectionInfo describes where the USB device can be reached.
type USBDeviceConnectionInfo struct {
	Host          string `json:"host,omitempty"`
	Port          int32  `json:"port,omitempty"`
	ExportedBusID string `json:"exportedBusID,omitempty"`
}

// USBDeviceStatus defines the observed state of USBDevice.
type USBDeviceStatus struct {
	Phase          string                   `json:"phase,omitempty"`
	ConnectionInfo *USBDeviceConnectionInfo `json:"connectionInfo,omitempty"`
	LastSeen       *metav1.Time             `json:"lastSeen,omitempty"`
	Health         string                   `json:"health,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=usbdev
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Node",type=string,JSONPath=`.spec.nodeName`

// USBDevice is the Schema for the usbdevices API.
type USBDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   USBDeviceSpec   `json:"spec,omitempty"`
	Status USBDeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// USBDeviceList contains a list of USBDevice.
type USBDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []USBDevice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&USBDevice{}, &USBDeviceList{})
}
