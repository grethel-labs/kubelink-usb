package agent

import (
	"context"
	"io"
	"log"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
	"github.com/grethel-labs/kubelink-usb/internal/utils"
)

func TestSerialFromPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "serial symlink path", path: "/dev/serial/by-id/usb-Example_Device_1234", want: "usb-Example_Device_1234"},
		{name: "tty path", path: "/dev/ttyUSB0", want: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := serialFromPath(tt.path); got != tt.want {
				t.Fatalf("serialFromPath(%q)=%q want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestMatchesRemovedPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		device *usbv1alpha1.USBDevice
		path   string
		serial string
		want   bool
	}{
		{
			name:   "exact device path",
			device: &usbv1alpha1.USBDevice{Spec: usbv1alpha1.USBDeviceSpec{DevicePath: "/dev/ttyUSB0"}},
			path:   "/dev/ttyUSB0",
			want:   true,
		},
		{
			name:   "serial match",
			device: &usbv1alpha1.USBDevice{Spec: usbv1alpha1.USBDeviceSpec{SerialNumber: "usb-abc"}},
			path:   "/dev/serial/by-id/usb-abc",
			serial: "usb-abc",
			want:   true,
		},
		{
			name:   "base path match",
			device: &usbv1alpha1.USBDevice{Spec: usbv1alpha1.USBDeviceSpec{DevicePath: "/dev/ttyUSB1"}},
			path:   "/other/ttyUSB1",
			want:   true,
		},
		{
			name:   "no match",
			device: &usbv1alpha1.USBDevice{Spec: usbv1alpha1.USBDeviceSpec{DevicePath: "/dev/ttyACM0", SerialNumber: "abc"}},
			path:   "/dev/ttyUSB1",
			serial: "different",
			want:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := matchesRemovedPath(tt.device, tt.path, tt.serial); got != tt.want {
				t.Fatalf("matchesRemovedPath()=%v want %v", got, tt.want)
			}
		})
	}
}

func TestBridgeAddAndRemove(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := usbv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme() error = %v", err)
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&usbv1alpha1.USBDevice{}).
		Build()

	bridge := NewUSBDeviceBridge(client, "node-a", log.New(io.Discard, "", 0))
	addPath := "/dev/serial/by-id/usb-Test_Device_001"
	if err := bridge.OnDiscoveryEvent(context.Background(), DiscoveryEvent{Type: DiscoveryEventAdd, Path: addPath}); err != nil {
		t.Fatalf("OnDiscoveryEvent(add) error = %v", err)
	}

	name := utils.DeviceFingerprint("node-a", "0000", "0000", "usb-Test_Device_001", "usb-Test_Device_001")
	var device usbv1alpha1.USBDevice
	if err := client.Get(context.Background(), types.NamespacedName{Name: name}, &device); err != nil {
		t.Fatalf("Get(created device) error = %v", err)
	}
	if device.Spec.DevicePath != addPath {
		t.Fatalf("device path=%q want %q", device.Spec.DevicePath, addPath)
	}

	device.Status.Phase = "Approved"
	if err := client.Status().Update(context.Background(), &device); err != nil {
		t.Fatalf("Status().Update() error = %v", err)
	}

	if err := bridge.OnDiscoveryEvent(context.Background(), DiscoveryEvent{Type: DiscoveryEventRemove, Path: addPath}); err != nil {
		t.Fatalf("OnDiscoveryEvent(remove) error = %v", err)
	}

	if err := client.Get(context.Background(), types.NamespacedName{Name: name}, &device); err != nil {
		t.Fatalf("Get(disconnected device) error = %v", err)
	}
	if device.Status.Phase != "Disconnected" {
		t.Fatalf("phase=%q want Disconnected", device.Status.Phase)
	}
	if device.Status.LastSeen == nil || device.Status.LastSeen.Equal(&metav1.Time{}) {
		t.Fatal("expected LastSeen to be set on remove")
	}
}
