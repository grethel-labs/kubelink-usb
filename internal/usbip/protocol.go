package usbip

import (
	"encoding/binary"
	"io"
)

const (
	OPReqDevList uint16 = 0x8005
	OPRepDevList uint16 = 0x0005
	OPReqImport  uint16 = 0x8003
	OPRepImport  uint16 = 0x0003

	USBIPCmdSubmit uint32 = 0x0001
	USBIPRetSubmit uint32 = 0x0003
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
