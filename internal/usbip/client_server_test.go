package usbip

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"
)

// mockProvider implements DeviceProvider for tests.
type mockProvider struct {
	devices []DeviceInfo
}

func (m *mockProvider) ListDevices() []DeviceInfo { return m.devices }
func (m *mockProvider) GetDevice(busID string) (*DeviceInfo, bool) {
	for _, d := range m.devices {
		if string(bytes.TrimRight(d.BusID[:], "\x00")) == busID {
			return &d, true
		}
	}
	return nil, false
}

func TestDevListRequestRoundtrip(t *testing.T) {
	t.Parallel()

	req := DevListRequest()
	var buf bytes.Buffer
	if err := req.Encode(&buf); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	var decoded BasicHeader
	if err := decoded.Decode(&buf); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if decoded.Code != OPReqDevList {
		t.Fatalf("Code = 0x%04x, want 0x%04x", decoded.Code, OPReqDevList)
	}
	if decoded.Version != USBIPVersion {
		t.Fatalf("Version = 0x%04x, want 0x%04x", decoded.Version, USBIPVersion)
	}
}

func TestDevListResponseRoundtrip(t *testing.T) {
	t.Parallel()

	var busID [32]byte
	copy(busID[:], "1-1")

	original := &DevListResponse{
		Devices: []DeviceInfo{
			{BusID: busID, BusNum: 1, DevNum: 2, Speed: 3, VendorID: 0x04b4, ProductID: 0x6001},
		},
	}

	var buf bytes.Buffer
	if err := original.Encode(&buf); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	decoded, err := DecodeDevListResponse(&buf)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if len(decoded.Devices) != 1 {
		t.Fatalf("Devices count = %d, want 1", len(decoded.Devices))
	}
	if decoded.Devices[0].VendorID != 0x04b4 {
		t.Fatalf("VendorID = 0x%04x, want 0x04b4", decoded.Devices[0].VendorID)
	}
}

func TestImportRequestRoundtrip(t *testing.T) {
	t.Parallel()

	req := NewImportRequest("1-1")
	var buf bytes.Buffer
	if err := req.Encode(&buf); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	decoded, err := DecodeImportRequest(&buf)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	busID := string(bytes.TrimRight(decoded.BusID[:], "\x00"))
	if busID != "1-1" {
		t.Fatalf("BusID = %q, want %q", busID, "1-1")
	}
}

func TestImportResponseRoundtrip(t *testing.T) {
	t.Parallel()

	var busID [32]byte
	copy(busID[:], "2-3")
	resp := &ImportResponse{
		BusID:     busID,
		BusNum:    2,
		DevNum:    3,
		Speed:     480,
		VendorID:  0x1234,
		ProductID: 0x5678,
	}

	var buf bytes.Buffer
	if err := resp.Encode(&buf); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	decoded, err := DecodeImportResponse(&buf)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if decoded.VendorID != 0x1234 {
		t.Fatalf("VendorID = 0x%04x, want 0x1234", decoded.VendorID)
	}
	if decoded.ProductID != 0x5678 {
		t.Fatalf("ProductID = 0x%04x, want 0x5678", decoded.ProductID)
	}
}

func TestBasicHeaderRoundtrip(t *testing.T) {
	t.Parallel()

	original := &BasicHeader{Version: USBIPVersion, Code: OPReqDevList, Status: 0}
	var buf bytes.Buffer
	if err := original.Encode(&buf); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	if buf.Len() != 8 {
		t.Fatalf("encoded size = %d want 8", buf.Len())
	}

	decoded := &BasicHeader{}
	if err := decoded.Decode(&buf); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if *decoded != *original {
		t.Fatalf("decoded %+v != original %+v", decoded, original)
	}
}

func TestBasicHeaderDecodeTruncated(t *testing.T) {
	t.Parallel()

	buf := bytes.NewBuffer([]byte{0x01, 0x02})
	h := &BasicHeader{}
	if err := h.Decode(buf); err == nil {
		t.Fatal("expected error on truncated input")
	}
}

func TestDecodeDevListResponseBadOpcode(t *testing.T) {
	t.Parallel()

	header := &BasicHeader{Version: USBIPVersion, Code: OPReqDevList, Status: 0}
	var buf bytes.Buffer
	header.Encode(&buf)
	// Write fake device count.
	buf.Write([]byte{0, 0, 0, 0})

	_, err := DecodeDevListResponse(&buf)
	if err == nil {
		t.Fatal("expected error for wrong opcode")
	}
}

func TestServerClientIntegration(t *testing.T) {
	t.Parallel()

	var busID [32]byte
	copy(busID[:], "1-1")
	provider := &mockProvider{
		devices: []DeviceInfo{
			{BusID: busID, VendorID: 0x04b4, ProductID: 0x6001},
		},
	}

	// Pick a free port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	addr := ln.Addr().String()
	ln.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start server in background.
	errCh := make(chan error, 1)
	go func() {
		errCh <- ServeWithProvider(ctx, addr, provider)
	}()

	// Brief wait for server to start.
	time.Sleep(50 * time.Millisecond)

	// Connect and list devices.
	devices, err := ListRemoteDevices(ctx, addr)
	if err != nil {
		t.Fatalf("ListRemoteDevices() error = %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("device count = %d, want 1", len(devices))
	}
	if devices[0].VendorID != 0x04b4 {
		t.Fatalf("VendorID = 0x%04x, want 0x04b4", devices[0].VendorID)
	}

	cancel()
}
