package usbip

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestBasicHeaderEncodeBigEndian(t *testing.T) {
	t.Parallel()

	header := BasicHeader{
		Version: 0x0111,
		Code:    OPReqImport,
		Status:  0x01020304,
	}

	var buf bytes.Buffer
	if err := header.Encode(&buf); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	want := []byte{
		0x01, 0x11, // Version
		0x80, 0x03, // Code
		0x01, 0x02, 0x03, 0x04, // Status
	}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("Encode() bytes = %v want %v", buf.Bytes(), want)
	}
}

func TestBasicHeaderDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	original := BasicHeader{
		Version: 0x0111,
		Code:    OPRepImport,
		Status:  0xAABBCCDD,
	}

	var buf bytes.Buffer
	if err := original.Encode(&buf); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	var decoded BasicHeader
	if err := decoded.Decode(&buf); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if decoded != original {
		t.Fatalf("roundtrip mismatch: got %+v want %+v", decoded, original)
	}
}

func TestBasicHeaderDecodeTruncatedInput(t *testing.T) {
	t.Parallel()

	var decoded BasicHeader
	err := decoded.Decode(bytes.NewReader([]byte{0x01, 0x11, 0x80}))
	if err == nil {
		t.Fatal("Decode() expected error for truncated input, got nil")
	}
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("Decode() error = %v want %v", err, io.ErrUnexpectedEOF)
	}
}
