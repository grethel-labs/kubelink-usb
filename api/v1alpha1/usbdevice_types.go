package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// USBDeviceSpec defines the desired state of USBDevice.
// It captures the physical identity of a USB device discovered on a cluster node,
// including vendor/product identifiers, bus path, and optional serial number.
//
// @relates USBDevice ||--|| USBDeviceSpec : "has spec"
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

// USBDeviceConnectionInfo describes where the USB device can be reached
// when it has been exported via USB/IP. The host:port pair identifies the
// usbipd endpoint and the exported bus ID names the virtual bus.
type USBDeviceConnectionInfo struct {
	Host          string `json:"host,omitempty"`
	Port          int32  `json:"port,omitempty"`
	ExportedBusID string `json:"exportedBusID,omitempty"`
}

// USBDeviceStatus defines the observed state of USBDevice.
// Phase tracks the device lifecycle from initial discovery through approval
// to active export or disconnection.
//
// @state [*] --> PendingApproval : Agent discovers device
// @state PendingApproval --> Approved : Approval granted
// @state PendingApproval --> Denied : Policy or manual reject
// @state Approved --> Exported : Agent exports via USB/IP
// @state Exported --> Disconnected : Device unplugged
// @state Disconnected --> PendingApproval : Reconnect (unknown)
// @state Disconnected --> Approved : Reconnect (whitelisted)
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
// It represents a physical USB device discovered on a Kubernetes node.
// Each device goes through a security-first approval workflow before
// it can be shared across the cluster via USB/IP tunnels.
//
// @component USBDevice["USBDevice CR"] --> USBDeviceApproval["Approval CR"]
// @component USBDevice --> USBConnection["Connection CR"]
// @relates USBDevice ||--o{ USBDeviceApproval : "referenced by"
// @relates USBDevice ||--o{ USBConnection : "referenced by"
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
