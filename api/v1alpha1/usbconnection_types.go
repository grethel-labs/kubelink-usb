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
// It references a cluster-scoped USBDevice and identifies the client node
// that should receive the USB/IP tunnel attachment.
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
// Phase tracks the tunnel lifecycle from request through establishment
// to eventual disconnection or failure.
//
// @state [*] --> Pending : Connection requested
// @state Pending --> Connecting : Device approved + export ready
// @state Connecting --> Connected : Attach successful
// @state Connected --> Disconnected : Network/device failure
// @state Disconnected --> Connecting : Reconnect attempt
// @state Disconnected --> Failed : Max retries exhausted
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
// It represents a USB/IP tunnel between a source node (device host) and
// a client node (device consumer). The tunnel lifecycle is managed by the
// USBConnectionReconciler and terminated via finalizer cleanup.
//
// @component USBConnection["Connection CR"] --> AgentServer["Agent Export"]
// @component USBConnection --> AgentClient["Agent Attach"]
// @relates USBConnection }o--|| USBDevice : "targets"
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
