package webhook

import (
	"context"
	"testing"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
)

func TestPolicyValidator(t *testing.T) {
	t.Parallel()

	validator := &PolicyValidator{}
	tests := []struct {
		name    string
		vendor  string
		product string
		wantErr bool
	}{
		{name: "valid IDs", vendor: "04b4", product: "6001", wantErr: false},
		{name: "invalid vendor", vendor: "XYZ", product: "6001", wantErr: true},
		{name: "invalid product", vendor: "04b4", product: "12", wantErr: true},
		{name: "empty IDs allowed", vendor: "", product: "", wantErr: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			obj := &usbv1alpha1.USBDevicePolicy{Spec: usbv1alpha1.USBDevicePolicySpec{Selector: usbv1alpha1.USBDevicePolicySelector{VendorID: tt.vendor, ProductID: tt.product}}}
			_, err := validator.ValidateCreate(context.Background(), obj)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestPolicyValidatorConstructorsAndNoopMethods(t *testing.T) {
	t.Parallel()

	validator := NewPolicyValidator()
	if validator == nil {
		t.Fatal("expected validator")
	}

	obj := &usbv1alpha1.USBDevicePolicy{}
	if _, err := validator.ValidateUpdate(context.Background(), obj, obj); err != nil {
		t.Fatalf("ValidateUpdate() unexpected error: %v", err)
	}
	if _, err := validator.ValidateDelete(context.Background(), obj); err != nil {
		t.Fatalf("ValidateDelete() unexpected error: %v", err)
	}
}

func TestDeviceDefaulter(t *testing.T) {
	t.Parallel()

	defaulter := &DeviceDefaulter{}
	tests := []struct {
		name      string
		device    *usbv1alpha1.USBDevice
		wantClass string
		wantDesc  string
		wantNode  string
	}{
		{
			name:      "fills all defaults",
			device:    &usbv1alpha1.USBDevice{},
			wantClass: "unknown",
			wantDesc:  "USB device",
			wantNode:  "unknown-node",
		},
		{
			name: "preserves existing values",
			device: &usbv1alpha1.USBDevice{Spec: usbv1alpha1.USBDeviceSpec{
				DeviceClass: "CDC",
				Description: "custom",
				NodeName:    "node-a",
			}},
			wantClass: "CDC",
			wantDesc:  "custom",
			wantNode:  "node-a",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := defaulter.Default(context.Background(), tt.device); err != nil {
				t.Fatalf("Default() error = %v", err)
			}
			if tt.device.Spec.DeviceClass != tt.wantClass {
				t.Fatalf("DeviceClass=%q want %q", tt.device.Spec.DeviceClass, tt.wantClass)
			}
			if tt.device.Spec.Description != tt.wantDesc {
				t.Fatalf("Description=%q want %q", tt.device.Spec.Description, tt.wantDesc)
			}
			if tt.device.Spec.NodeName != tt.wantNode {
				t.Fatalf("NodeName=%q want %q", tt.device.Spec.NodeName, tt.wantNode)
			}
		})
	}
}

func TestDeviceDefaulterConstructor(t *testing.T) {
	t.Parallel()

	defaulter := NewDeviceDefaulter()
	if defaulter == nil {
		t.Fatal("expected defaulter")
	}
}
