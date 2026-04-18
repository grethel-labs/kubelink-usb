package utils

import (
	"net"
	"testing"
)

func TestIsTCPReachable(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen() error = %v", err)
	}
	addr := ln.Addr().String()

	if !IsTCPReachable(addr) {
		t.Fatalf("IsTCPReachable(%q)=false want true", addr)
	}

	if err := ln.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if IsTCPReachable(addr) {
		t.Fatalf("IsTCPReachable(%q)=true want false after listener close", addr)
	}
}
