package usbip

import (
	"context"
	"fmt"
	"net"
)

// Connect negotiates an import session with a USB/IP server.
// Used by the connection controller to establish cross-node USB device access.
//
// Intent: Provide a single entrypoint for device list retrieval from a remote server.
// Inputs: Context for cancellation and server address.
// Outputs: Nil on success.
// Errors: Returns connection or protocol errors.
func Connect(ctx context.Context, addr string) error {
	_, err := ListRemoteDevices(ctx, addr)
	return err
}

// ListRemoteDevices connects to a USB/IP server and retrieves the exported device list.
// Called during health monitoring to verify remote device availability.
//
// Intent: Fetch available USB devices from a remote node's USB/IP export daemon.
// Inputs: Context for cancellation and server address (host:port).
// Outputs: Slice of DeviceInfo describing exported devices.
// Errors: Returns dial, encode, or decode errors.
func ListRemoteDevices(ctx context.Context, addr string) ([]DeviceInfo, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}
	defer conn.Close()

	req := DevListRequest()
	if err := req.Encode(conn); err != nil {
		return nil, fmt.Errorf("send devlist request: %w", err)
	}

	resp, err := DecodeDevListResponse(conn)
	if err != nil {
		return nil, fmt.Errorf("decode devlist response: %w", err)
	}

	return resp.Devices, nil
}

// ImportDevice connects to a USB/IP server and imports a specific device by bus ID.
// Called by the connection controller after approval to bind a remote device locally.
//
// Intent: Attach a remote USB device so the local node can use it.
// Inputs: Context, server address, and the bus ID of the device to import.
// Outputs: ImportResponse with device info on success.
// Errors: Returns dial, protocol, or rejection errors.
func ImportDevice(ctx context.Context, addr, busID string) (*ImportResponse, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}
	defer conn.Close()

	req := NewImportRequest(busID)
	if err := req.Encode(conn); err != nil {
		return nil, fmt.Errorf("send import request: %w", err)
	}

	resp, err := DecodeImportResponse(conn)
	if err != nil {
		return nil, fmt.Errorf("decode import response: %w", err)
	}

	if resp.Header.Status != 0 {
		return nil, fmt.Errorf("import rejected with status %d", resp.Header.Status)
	}

	return resp, nil
}
