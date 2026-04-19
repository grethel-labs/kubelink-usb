package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// USBConnectionDeviceRef references the target cluster-scoped USBDevice.
type USBConnectionDeviceRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}

// USBConnectionPodSelector identifies workloads that consume the connection.
type USBConnectionPodSelector struct {
	MatchLabels map[string]string `json:"matchLabels,omitempty"`
}

// USBConnectionSpec defines the desired state of USBConnection.
type USBConnectionSpec struct {
	DeviceRef   USBConnectionDeviceRef    `json:"deviceRef"`
	ClientNode  string                    `json:"clientNode"`
	PodSelector *USBConnectionPodSelector `json:"podSelector,omitempty"`
}

// USBConnectionTunnelInfo captures active tunnel details.
type USBConnectionTunnelInfo struct {
	ServerHost string `json:"serverHost,omitempty"`
	ServerPort int32  `json:"serverPort,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
}

// USBConnectionStatus defines the observed state of USBConnection.
type USBConnectionStatus struct {
	Phase            string                   `json:"phase,omitempty"`
	ClientDevicePath string                   `json:"clientDevicePath,omitempty"`
	RetryCount       int32                    `json:"retryCount,omitempty"`
	LastRetryTime    *metav1.Time             `json:"lastRetryTime,omitempty"`
	TunnelInfo       *USBConnectionTunnelInfo `json:"tunnelInfo,omitempty"`
	Conditions       []metav1.Condition       `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=usbconn

// USBConnection is the Schema for the usbconnections API.
type USBConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   USBConnectionSpec   `json:"spec,omitempty"`
	Status USBConnectionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// USBConnectionList contains a list of USBConnection.
type USBConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []USBConnection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&USBConnection{}, &USBConnectionList{})
}
