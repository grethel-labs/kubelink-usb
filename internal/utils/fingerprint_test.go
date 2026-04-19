package utils

import (
	"strings"
	"testing"
)

func TestDeviceFingerprint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		nodeName     string
		vendorID     string
		productID    string
		serialNumber string
		busID        string
		want         string
	}{
		{
			name:         "normal with serial",
			nodeName:     "node-a",
			vendorID:     "04b4",
			productID:    "6001",
			serialNumber: "ABC123",
			busID:        "1-1",
			want:         "node-a-04b4-6001-abc123",
		},
		{
			name:         "falls back to busID when no serial",
			nodeName:     "node-b",
			vendorID:     "1234",
			productID:    "5678",
			serialNumber: "",
			busID:        "2-3.1",
			want:         "node-b-1234-5678-2-3-1",
		},
		{
			name:         "special characters sanitized",
			nodeName:     "Node_A",
			vendorID:     "04B4",
			productID:    "6001",
			serialNumber: "SER!@#$%123",
			busID:        "1-1",
			want:         "node-a-04b4-6001-ser-123",
		},
		{
			name:         "all empty fields",
			nodeName:     "",
			vendorID:     "",
			productID:    "",
			serialNumber: "",
			busID:        "",
			want:         "unknown-device",
		},
		{
			name:         "truncates to 63 chars",
			nodeName:     "very-long-node-name-that-goes-on-forever",
			vendorID:     "abcd",
			productID:    "ef01",
			serialNumber: "serial-number-that-is-also-extremely-long-for-testing",
			busID:        "1-1",
			want:         "very-long-node-name-that-goes-on-forever-abcd-ef01-serial-numbe",
		},
		{
			name:         "leading/trailing dashes stripped",
			nodeName:     "-node-",
			vendorID:     "-vid-",
			productID:    "-pid-",
			serialNumber: "-ser-",
			busID:        "1-1",
			want:         "node-vid-pid-ser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := DeviceFingerprint(tt.nodeName, tt.vendorID, tt.productID, tt.serialNumber, tt.busID)
			if got != tt.want {
				t.Errorf("DeviceFingerprint() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSanitizeDNSLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  string
	}{
		{"hello-world", "hello-world"},
		{"UPPER-CASE", "upper-case"},
		{"special!@#chars", "special-chars"},
		{"---leading---trailing---", "leading-trailing"},
		{"", "unknown-device"},
		{strings.Repeat("a", 100), strings.Repeat("a", 63)},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := sanitizeDNSLabel(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeDNSLabel(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
