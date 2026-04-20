package webhook

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
)

// DeviceDefaulter sets defaults for USBDevice fields.
type DeviceDefaulter struct{}

// NewDeviceDefaulter creates a USBDevice defaulter.
func NewDeviceDefaulter() admission.CustomDefaulter {
	return &DeviceDefaulter{}
}

func (d *DeviceDefaulter) Default(_ context.Context, obj runtime.Object) error {
	device, ok := obj.(*usbv1alpha1.USBDevice)
	if !ok {
		return fmt.Errorf("expected USBDevice, got %T", obj)
	}
	if device.Spec.DeviceClass == "" {
		device.Spec.DeviceClass = "unknown"
	}
	if device.Spec.Description == "" {
		device.Spec.Description = "USB device"
	}
	if device.Spec.NodeName == "" {
		device.Spec.NodeName = "unknown-node"
	}
	return nil
}
