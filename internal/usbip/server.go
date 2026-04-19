package usbip

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"net"
)

// DeviceProvider supplies the list of exported devices.
type DeviceProvider interface {
	ListDevices() []DeviceInfo
	GetDevice(busID string) (*DeviceInfo, bool)
}

// Serve starts the USB/IP export serving loop on the given bind address.
//
// Intent: Provide a stable interface for export daemon lifecycle.
// Inputs: Context for cancellation, bind address, and optional device provider.
// Outputs: Returns when context is cancelled or an error occurs.
// Errors: Returns listener or connection handling errors.
func Serve(ctx context.Context, addr string) error {
	return ServeWithProvider(ctx, addr, nil)
}

// ServeWithProvider starts a USB/IP server with a custom device provider.
func ServeWithProvider(ctx context.Context, addr string, provider DeviceProvider) error {
	lc := net.ListenConfig{}
	ln, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}
	defer ln.Close()

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				return fmt.Errorf("accept: %w", err)
			}
		}
		go handleConnection(ctx, conn, provider)
	}
}

func handleConnection(ctx context.Context, conn net.Conn, provider DeviceProvider) {
	defer conn.Close()

	var header BasicHeader
	if err := header.Decode(conn); err != nil {
		return
	}

	switch header.Code {
	case OPReqDevList:
		handleDevList(conn, provider)
	case OPReqImport:
		handleImport(conn, provider)
	}
}

func handleDevList(conn net.Conn, provider DeviceProvider) {
	var devices []DeviceInfo
	if provider != nil {
		devices = provider.ListDevices()
	}
	resp := &DevListResponse{
		Header:  BasicHeader{Version: USBIPVersion, Code: OPRepDevList, Status: 0},
		Devices: devices,
	}
	resp.Encode(conn)
}

func handleImport(conn net.Conn, provider DeviceProvider) {
	var busID [32]byte
	if err := binary.Read(conn, binary.BigEndian, &busID); err != nil {
		return
	}

	resp := &ImportResponse{
		Header: BasicHeader{Version: USBIPVersion, Code: OPRepImport, Status: 0},
		BusID:  busID,
	}

	if provider != nil {
		busIDStr := string(bytes.TrimRight(busID[:], "\x00"))
		if dev, ok := provider.GetDevice(busIDStr); ok {
			resp.BusNum = dev.BusNum
			resp.DevNum = dev.DevNum
			resp.Speed = dev.Speed
			resp.VendorID = dev.VendorID
			resp.ProductID = dev.ProductID
		} else {
			resp.Header.Status = 1
		}
	}

	resp.Encode(conn)
}
