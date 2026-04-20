package usbip

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	OPReqDevList uint16 = 0x8005
	OPRepDevList uint16 = 0x0005
	OPReqImport  uint16 = 0x8003
	OPRepImport  uint16 = 0x0003

	USBIPCmdSubmit uint32 = 0x0001
	USBIPRetSubmit uint32 = 0x0003

	// USBIPVersion is the standard protocol version header.
	USBIPVersion uint16 = 0x0111
)

// BasicHeader mirrors the USB/IP operation header in big-endian encoding.
type BasicHeader struct {
	Version uint16
	Code    uint16
	Status  uint32
}

func (h *BasicHeader) Encode(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, h)
}

func (h *BasicHeader) Decode(r io.Reader) error {
	return binary.Read(r, binary.BigEndian, h)
}

// DevListRequest creates a device list request header.
func DevListRequest() *BasicHeader {
	return &BasicHeader{
		Version: USBIPVersion,
		Code:    OPReqDevList,
		Status:  0,
	}
}

// DevListResponse represents a device list response with device info.
type DevListResponse struct {
	Header  BasicHeader
	Devices []DeviceInfo
}

// DeviceInfo describes a single exported USB device.
type DeviceInfo struct {
	BusID     [32]byte
	BusNum    uint32
	DevNum    uint32
	Speed     uint32
	VendorID  uint16
	ProductID uint16
}

// Encode writes the device list response to a writer.
func (r *DevListResponse) Encode(w io.Writer) error {
	r.Header.Version = USBIPVersion
	r.Header.Code = OPRepDevList
	if err := r.Header.Encode(w); err != nil {
		return fmt.Errorf("encode header: %w", err)
	}
	numDevices := uint32(len(r.Devices))
	if err := binary.Write(w, binary.BigEndian, numDevices); err != nil {
		return fmt.Errorf("encode device count: %w", err)
	}
	for _, dev := range r.Devices {
		if err := binary.Write(w, binary.BigEndian, &dev); err != nil {
			return fmt.Errorf("encode device: %w", err)
		}
	}
	return nil
}

// DecodeDevListResponse reads a device list response from a reader.
func DecodeDevListResponse(r io.Reader) (*DevListResponse, error) {
	resp := &DevListResponse{}
	if err := resp.Header.Decode(r); err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}
	if resp.Header.Code != OPRepDevList {
		return nil, fmt.Errorf("unexpected opcode: 0x%04x", resp.Header.Code)
	}
	var numDevices uint32
	if err := binary.Read(r, binary.BigEndian, &numDevices); err != nil {
		return nil, fmt.Errorf("decode device count: %w", err)
	}
	if numDevices > 256 {
		return nil, fmt.Errorf("device count %d exceeds maximum 256", numDevices)
	}
	resp.Devices = make([]DeviceInfo, numDevices)
	for i := range resp.Devices {
		if err := binary.Read(r, binary.BigEndian, &resp.Devices[i]); err != nil {
			return nil, fmt.Errorf("decode device %d: %w", i, err)
		}
	}
	return resp, nil
}

// ImportRequest creates an import request for a specific bus ID.
type ImportRequest struct {
	Header BasicHeader
	BusID  [32]byte
}

// NewImportRequest creates an import request.
func NewImportRequest(busID string) *ImportRequest {
	req := &ImportRequest{
		Header: BasicHeader{
			Version: USBIPVersion,
			Code:    OPReqImport,
			Status:  0,
		},
	}
	copy(req.BusID[:], busID)
	return req
}

// Encode writes the import request.
func (r *ImportRequest) Encode(w io.Writer) error {
	if err := r.Header.Encode(w); err != nil {
		return fmt.Errorf("encode header: %w", err)
	}
	return binary.Write(w, binary.BigEndian, r.BusID)
}

// DecodeImportRequest reads an import request.
func DecodeImportRequest(r io.Reader) (*ImportRequest, error) {
	req := &ImportRequest{}
	if err := req.Header.Decode(r); err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}
	if req.Header.Code != OPReqImport {
		return nil, fmt.Errorf("unexpected opcode: 0x%04x", req.Header.Code)
	}
	if err := binary.Read(r, binary.BigEndian, &req.BusID); err != nil {
		return nil, fmt.Errorf("decode busID: %w", err)
	}
	return req, nil
}

// ImportResponse represents the response to an import request.
type ImportResponse struct {
	Header    BasicHeader
	BusID     [32]byte
	BusNum    uint32
	DevNum    uint32
	Speed     uint32
	VendorID  uint16
	ProductID uint16
}

// Encode writes the import response.
func (r *ImportResponse) Encode(w io.Writer) error {
	r.Header.Version = USBIPVersion
	r.Header.Code = OPRepImport
	if err := r.Header.Encode(w); err != nil {
		return fmt.Errorf("encode header: %w", err)
	}
	if err := binary.Write(w, binary.BigEndian, r.BusID); err != nil {
		return fmt.Errorf("encode busID: %w", err)
	}
	if err := binary.Write(w, binary.BigEndian, r.BusNum); err != nil {
		return fmt.Errorf("encode busNum: %w", err)
	}
	if err := binary.Write(w, binary.BigEndian, r.DevNum); err != nil {
		return fmt.Errorf("encode devNum: %w", err)
	}
	if err := binary.Write(w, binary.BigEndian, r.Speed); err != nil {
		return fmt.Errorf("encode speed: %w", err)
	}
	if err := binary.Write(w, binary.BigEndian, r.VendorID); err != nil {
		return fmt.Errorf("encode vendorID: %w", err)
	}
	return binary.Write(w, binary.BigEndian, r.ProductID)
}

// DecodeImportResponse reads an import response.
func DecodeImportResponse(r io.Reader) (*ImportResponse, error) {
	resp := &ImportResponse{}
	if err := resp.Header.Decode(r); err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}
	if resp.Header.Code != OPRepImport {
		return nil, fmt.Errorf("unexpected opcode: 0x%04x", resp.Header.Code)
	}
	if err := binary.Read(r, binary.BigEndian, &resp.BusID); err != nil {
		return nil, fmt.Errorf("decode busID: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &resp.BusNum); err != nil {
		return nil, fmt.Errorf("decode busNum: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &resp.DevNum); err != nil {
		return nil, fmt.Errorf("decode devNum: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &resp.Speed); err != nil {
		return nil, fmt.Errorf("decode speed: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &resp.VendorID); err != nil {
		return nil, fmt.Errorf("decode vendorID: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &resp.ProductID); err != nil {
		return nil, fmt.Errorf("decode productID: %w", err)
	}
	return resp, nil
}
