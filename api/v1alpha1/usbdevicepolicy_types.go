package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// USBDevicePolicySelector identifies which devices are targeted by this policy.
type USBDevicePolicySelector struct {
	VendorID  string   `json:"vendorID,omitempty"`
	ProductID string   `json:"productID,omitempty"`
	NodeNames []string `json:"nodeNames,omitempty"`
}

// USBDeviceApprovalConfig configures approval behavior.
type USBDeviceApprovalConfig struct {
	Mode                    string          `json:"mode,omitempty"`
	AutoApproveKnownDevices bool            `json:"autoApproveKnownDevices,omitempty"`
	ApprovalTimeout         metav1.Duration `json:"approvalTimeout,omitempty"`
}

// USBDeviceRestrictions defines policy restrictions.
type USBDeviceRestrictions struct {
	AllowedNodes              []string `json:"allowedNodes,omitempty"`
	AllowedNamespaces         []string `json:"allowedNamespaces,omitempty"`
	MaxConcurrentConnections  int32    `json:"maxConcurrentConnections,omitempty"`
	Readonly                  bool     `json:"readonly,omitempty"`
	RequireEncryption         bool     `json:"requireEncryption,omitempty"`
	NetworkIsolation          bool     `json:"networkIsolation,omitempty"`
	AllowedDeviceClasses      []string `json:"allowedDeviceClasses,omitempty"`
	DenyHumanInterfaceDevices bool     `json:"denyHumanInterfaceDevices,omitempty"`
}

// USBDeviceLifecycle defines reconnect/disconnect controls.
type USBDeviceLifecycle struct {
	DisconnectTimeout metav1.Duration `json:"disconnectTimeout,omitempty"`
	ReconnectAttempts int32           `json:"reconnectAttempts,omitempty"`
	ReconnectBackoff  metav1.Duration `json:"reconnectBackoff,omitempty"`
}

// USBDevicePolicySpec defines the desired state of USBDevicePolicy.
type USBDevicePolicySpec struct {
	Selector     USBDevicePolicySelector `json:"selector,omitempty"`
	Approval     USBDeviceApprovalConfig `json:"approval,omitempty"`
	Restrictions USBDeviceRestrictions   `json:"restrictions,omitempty"`
	Lifecycle    USBDeviceLifecycle      `json:"lifecycle,omitempty"`
}

// USBDevicePolicyStatus defines the observed state of USBDevicePolicy.
type USBDevicePolicyStatus struct {
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=usbp

// USBDevicePolicy is the Schema for the usbdevicepolicies API.
type USBDevicePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   USBDevicePolicySpec   `json:"spec,omitempty"`
	Status USBDevicePolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// USBDevicePolicyList contains a list of USBDevicePolicy.
type USBDevicePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []USBDevicePolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&USBDevicePolicy{}, &USBDevicePolicyList{})
}
