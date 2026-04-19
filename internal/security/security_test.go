package security

import (
	"crypto/tls"
	"testing"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestWhitelistAddAndHas(t *testing.T) {
	t.Parallel()

	w := NewWhitelist()
	if w.Has("device-a") {
		t.Fatal("expected device-a to be absent before Add")
	}

	w.Add("device-a")
	if !w.Has("device-a") {
		t.Fatal("expected device-a to be present after Add")
	}
}

func TestTLSConfigMinVersionTLS13(t *testing.T) {
	t.Parallel()

	cfg := TLSConfig()
	if cfg == nil {
		t.Fatal("expected non-nil TLS config")
	}
	if cfg.MinVersion != tls.VersionTLS13 {
		t.Fatalf("TLSConfig().MinVersion=%d want %d", cfg.MinVersion, tls.VersionTLS13)
	}
}

func TestEngineAllows(t *testing.T) {
	t.Parallel()

	engine := &Engine{}

	tests := []struct {
		name   string
		device usbv1alpha1.USBDevice
		policy usbv1alpha1.USBDevicePolicy
		want   bool
	}{
		{
			name: "empty policy allows any device",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{VendorID: "04b4", ProductID: "6001", NodeName: "node-a"},
			},
			policy: usbv1alpha1.USBDevicePolicy{},
			want:   true,
		},
		{
			name: "vendor ID match",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{VendorID: "04b4", ProductID: "6001"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Selector: usbv1alpha1.USBDevicePolicySelector{VendorID: "04b4"},
				},
			},
			want: true,
		},
		{
			name: "vendor ID mismatch",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{VendorID: "1234"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Selector: usbv1alpha1.USBDevicePolicySelector{VendorID: "04b4"},
				},
			},
			want: false,
		},
		{
			name: "product ID mismatch",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{VendorID: "04b4", ProductID: "9999"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Selector: usbv1alpha1.USBDevicePolicySelector{VendorID: "04b4", ProductID: "6001"},
				},
			},
			want: false,
		},
		{
			name: "device on allowed node",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{NodeName: "node-a"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Restrictions: usbv1alpha1.USBDeviceRestrictions{AllowedNodes: []string{"node-a", "node-b"}},
				},
			},
			want: true,
		},
		{
			name: "device on denied node",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{NodeName: "node-c"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Restrictions: usbv1alpha1.USBDeviceRestrictions{AllowedNodes: []string{"node-a", "node-b"}},
				},
			},
			want: false,
		},
		{
			name: "node selector mismatch",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{NodeName: "node-c"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Selector: usbv1alpha1.USBDevicePolicySelector{NodeNames: []string{"node-a"}},
				},
			},
			want: false,
		},
		{
			name: "HID device denied",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{DeviceClass: "HID"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Restrictions: usbv1alpha1.USBDeviceRestrictions{DenyHumanInterfaceDevices: true},
				},
			},
			want: false,
		},
		{
			name: "HID device class 03 denied",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{DeviceClass: "03"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Restrictions: usbv1alpha1.USBDeviceRestrictions{DenyHumanInterfaceDevices: true},
				},
			},
			want: false,
		},
		{
			name: "non-HID device passes HID check",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{DeviceClass: "CDC"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Restrictions: usbv1alpha1.USBDeviceRestrictions{DenyHumanInterfaceDevices: true},
				},
			},
			want: true,
		},
		{
			name: "allowed device class",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{DeviceClass: "CDC"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Restrictions: usbv1alpha1.USBDeviceRestrictions{AllowedDeviceClasses: []string{"CDC", "vendor-specific"}},
				},
			},
			want: true,
		},
		{
			name: "disallowed device class",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{DeviceClass: "mass-storage"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Restrictions: usbv1alpha1.USBDeviceRestrictions{AllowedDeviceClasses: []string{"CDC"}},
				},
			},
			want: false,
		},
		{
			name: "case-insensitive vendor match",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{VendorID: "04B4"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Selector: usbv1alpha1.USBDevicePolicySelector{VendorID: "04b4"},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := engine.Allows(tt.device, tt.policy)
			if got != tt.want {
				t.Errorf("Allows() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEngineMatchesSelector(t *testing.T) {
	t.Parallel()

	engine := &Engine{}

	device := usbv1alpha1.USBDevice{
		ObjectMeta: metav1.ObjectMeta{Name: "dev1"},
		Spec:       usbv1alpha1.USBDeviceSpec{VendorID: "04b4", ProductID: "6001", NodeName: "node-a"},
	}

	matchingPolicy := usbv1alpha1.USBDevicePolicy{
		Spec: usbv1alpha1.USBDevicePolicySpec{
			Selector: usbv1alpha1.USBDevicePolicySelector{VendorID: "04b4"},
		},
	}
	nonMatchingPolicy := usbv1alpha1.USBDevicePolicy{
		Spec: usbv1alpha1.USBDevicePolicySpec{
			Selector: usbv1alpha1.USBDevicePolicySelector{VendorID: "dead"},
		},
	}

	if !engine.MatchesSelector(device, matchingPolicy) {
		t.Error("expected policy to match device")
	}
	if engine.MatchesSelector(device, nonMatchingPolicy) {
		t.Error("expected policy NOT to match device")
	}
}
