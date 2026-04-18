package agent

import (
	"context"
	"testing"
)

func TestClientAttachDetachCurrentStubBehavior(t *testing.T) {
	t.Parallel()

	c := &Client{}
	devicePath, err := c.Attach(context.Background(), "10.0.0.2:3240", "1-1")
	if err != nil {
		t.Fatalf("Attach() error = %v", err)
	}
	if devicePath != "" {
		t.Fatalf("Attach() path = %q want empty string for current stub", devicePath)
	}
	if err := c.Detach(context.Background(), "port-00"); err != nil {
		t.Fatalf("Detach() error = %v", err)
	}
}

func TestServerExportUnexportCurrentStubBehavior(t *testing.T) {
	t.Parallel()

	s := &Server{}
	if err := s.Export(context.Background(), "/dev/ttyUSB0"); err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	if err := s.Unexport(context.Background(), "/dev/ttyUSB0"); err != nil {
		t.Fatalf("Unexport() error = %v", err)
	}
}
