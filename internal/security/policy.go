package security

import usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"

// Engine evaluates policy decisions for discovered devices.
type Engine struct{}

func (e *Engine) Allows(_ usbv1alpha1.USBDevice, _ usbv1alpha1.USBDevicePolicy) bool { return true }
