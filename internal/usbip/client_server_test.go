package usbip

import (
	"context"
	"testing"
)

func TestConnectCurrentStubBehavior(t *testing.T) {
	t.Parallel()

	if err := Connect(context.Background(), "127.0.0.1:3240"); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
}

func TestServeCurrentStubBehavior(t *testing.T) {
	t.Parallel()

	if err := Serve(context.Background(), "127.0.0.1:3240"); err != nil {
		t.Fatalf("Serve() error = %v", err)
	}
}
