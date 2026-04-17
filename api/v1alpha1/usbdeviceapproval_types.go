package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// USBDeviceApprovalDeviceRef references the requested USB device.
type USBDeviceApprovalDeviceRef struct {
	Name string `json:"name"`
}

// USBDeviceApprovalPolicyRef references the policy that governs this approval.
type USBDeviceApprovalPolicyRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// USBDeviceApprovalSpec defines the desired state of USBDeviceApproval.
type USBDeviceApprovalSpec struct {
	DeviceRef      USBDeviceApprovalDeviceRef  `json:"deviceRef"`
	Requester      string                      `json:"requester"`
	PolicyRef      *USBDeviceApprovalPolicyRef `json:"policyRef,omitempty"`
	Phase          string                      `json:"phase,omitempty"`
	ApprovedBy     string                      `json:"approvedBy,omitempty"`
	ApprovedAt     *metav1.Time                `json:"approvedAt,omitempty"`
	ExpiresAt      *metav1.Time                `json:"expiresAt,omitempty"`
	DecisionReason string                      `json:"decisionReason,omitempty"`
}

// USBDeviceApprovalStatus defines observed approval state.
type USBDeviceApprovalStatus struct {
	Phase          string       `json:"phase,omitempty"`
	ApprovedBy     string       `json:"approvedBy,omitempty"`
	ApprovedAt     *metav1.Time `json:"approvedAt,omitempty"`
	ExpiresAt      *metav1.Time `json:"expiresAt,omitempty"`
	DecisionReason string       `json:"decisionReason,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=usbapproval

// USBDeviceApproval is the Schema for the usbdeviceapprovals API.
type USBDeviceApproval struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   USBDeviceApprovalSpec   `json:"spec,omitempty"`
	Status USBDeviceApprovalStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// USBDeviceApprovalList contains a list of USBDeviceApproval.
type USBDeviceApprovalList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []USBDeviceApproval `json:"items"`
}

func init() {
	SchemeBuilder.Register(&USBDeviceApproval{}, &USBDeviceApprovalList{})
}
