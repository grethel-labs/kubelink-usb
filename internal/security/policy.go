package security

import (
	"strings"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
)

// Engine evaluates policy decisions for discovered devices.
type Engine struct{}

// Allows checks whether a device is permitted by the given policy.
//
// Intent: Evaluate policy selectors and restrictions against discovered device attributes.
// Inputs: Device and policy to evaluate.
// Outputs: True if the device passes all policy checks, false otherwise.
// Errors: None.
func (e *Engine) Allows(device usbv1alpha1.USBDevice, policy usbv1alpha1.USBDevicePolicy) bool {
	sel := policy.Spec.Selector
	res := policy.Spec.Restrictions

	// Check vendor ID selector.
	if sel.VendorID != "" && !strings.EqualFold(sel.VendorID, device.Spec.VendorID) {
		return false
	}

	// Check product ID selector.
	if sel.ProductID != "" && !strings.EqualFold(sel.ProductID, device.Spec.ProductID) {
		return false
	}

	// Check node name selector.
	if len(sel.NodeNames) > 0 && !containsIgnoreCase(sel.NodeNames, device.Spec.NodeName) {
		return false
	}

	// Check allowed nodes restriction.
	if len(res.AllowedNodes) > 0 && !containsIgnoreCase(res.AllowedNodes, device.Spec.NodeName) {
		return false
	}

	// Check allowed device classes restriction.
	if len(res.AllowedDeviceClasses) > 0 && device.Spec.DeviceClass != "" {
		if !containsIgnoreCase(res.AllowedDeviceClasses, device.Spec.DeviceClass) {
			return false
		}
	}

	// Check HID denial.
	if res.DenyHumanInterfaceDevices && isHIDClass(device.Spec.DeviceClass) {
		return false
	}

	return true
}

// MatchesSelector returns true if the device matches the policy's selector fields only.
//
// Intent: Allow callers to determine if a policy applies to a device before evaluating restrictions.
// Inputs: Device and policy.
// Outputs: True if all non-empty selector fields match.
func (e *Engine) MatchesSelector(device usbv1alpha1.USBDevice, policy usbv1alpha1.USBDevicePolicy) bool {
	sel := policy.Spec.Selector

	if sel.VendorID != "" && !strings.EqualFold(sel.VendorID, device.Spec.VendorID) {
		return false
	}
	if sel.ProductID != "" && !strings.EqualFold(sel.ProductID, device.Spec.ProductID) {
		return false
	}
	if len(sel.NodeNames) > 0 && !containsIgnoreCase(sel.NodeNames, device.Spec.NodeName) {
		return false
	}
	return true
}

func containsIgnoreCase(items []string, target string) bool {
	for _, item := range items {
		if strings.EqualFold(item, target) {
			return true
		}
	}
	return false
}

func isHIDClass(deviceClass string) bool {
	lower := strings.ToLower(deviceClass)
	return lower == "hid" || lower == "03" || lower == "0x03"
}
