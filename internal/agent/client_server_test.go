package agent

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// mockRunner simulates command execution for tests.
type mockRunner struct {
	output []byte
	err    error
}

func (m *mockRunner) Run(_ context.Context, _ string, _ ...string) ([]byte, error) {
	return m.output, m.err
}

func TestServerExportSuccess(t *testing.T) {
	t.Parallel()

	s := &Server{Runner: &mockRunner{output: []byte("bound\n")}}
	if err := s.Export(context.Background(), "1-1"); err != nil {
		t.Fatalf("Export() error = %v", err)
	}
}

func TestServerExportEmptyBusID(t *testing.T) {
	t.Parallel()

	s := &Server{Runner: &mockRunner{}}
	if err := s.Export(context.Background(), ""); err == nil {
		t.Fatal("Export() expected error for empty busID")
	}
}

func TestServerExportCommandFailure(t *testing.T) {
	t.Parallel()

	s := &Server{Runner: &mockRunner{output: []byte("not found"), err: fmt.Errorf("exit status 1")}}
	err := s.Export(context.Background(), "1-1")
	if err == nil {
		t.Fatal("Export() expected error on command failure")
	}
}

func TestServerUnexportSuccess(t *testing.T) {
	t.Parallel()

	s := &Server{Runner: &mockRunner{output: []byte("unbound\n")}}
	if err := s.Unexport(context.Background(), "1-1"); err != nil {
		t.Fatalf("Unexport() error = %v", err)
	}
}

func TestServerUnexportEmptyBusID(t *testing.T) {
	t.Parallel()

	s := &Server{Runner: &mockRunner{}}
	if err := s.Unexport(context.Background(), ""); err == nil {
		t.Fatal("Unexport() expected error for empty busID")
	}
}

func TestServerRunnerDefault(t *testing.T) {
	t.Parallel()

	s := &Server{}
	if _, ok := s.runner().(*ExecRunner); !ok {
		t.Fatal("expected default runner to be ExecRunner")
	}
}

func TestExecRunnerRunSuccess(t *testing.T) {
	t.Parallel()

	r := &ExecRunner{}
	out, err := r.Run(context.Background(), "sh", "-c", "printf test-output")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if strings.TrimSpace(string(out)) != "test-output" {
		t.Fatalf("Run() output = %q", string(out))
	}
}

func TestClientAttachSuccess(t *testing.T) {
	t.Parallel()

	c := &Client{Runner: &mockRunner{output: []byte("/dev/ttyUSB0\n")}}
	path, err := c.Attach(context.Background(), "10.0.0.2:3240", "1-1")
	if err != nil {
		t.Fatalf("Attach() error = %v", err)
	}
	if path != "/dev/ttyUSB0" {
		t.Fatalf("Attach() path = %q want %q", path, "/dev/ttyUSB0")
	}
}

func TestClientAttachEmptyRemote(t *testing.T) {
	t.Parallel()

	c := &Client{Runner: &mockRunner{}}
	_, err := c.Attach(context.Background(), "", "1-1")
	if err == nil {
		t.Fatal("Attach() expected error for empty remote")
	}
}

func TestClientAttachEmptyBusID(t *testing.T) {
	t.Parallel()

	c := &Client{Runner: &mockRunner{}}
	_, err := c.Attach(context.Background(), "10.0.0.2:3240", "")
	if err == nil {
		t.Fatal("Attach() expected error for empty busID")
	}
}

func TestClientAttachCommandFailure(t *testing.T) {
	t.Parallel()

	c := &Client{Runner: &mockRunner{err: fmt.Errorf("exit status 1")}}
	_, err := c.Attach(context.Background(), "10.0.0.2:3240", "1-1")
	if err == nil {
		t.Fatal("Attach() expected error on command failure")
	}
}

func TestClientAttachNoDevicePath(t *testing.T) {
	t.Parallel()

	c := &Client{Runner: &mockRunner{output: []byte("attached to port 00\n")}}
	path, err := c.Attach(context.Background(), "10.0.0.2:3240", "1-1")
	if err != nil {
		t.Fatalf("Attach() error = %v", err)
	}
	if path != "" {
		t.Fatalf("Attach() path = %q want empty for non-matching output", path)
	}
}

func TestClientDetachSuccess(t *testing.T) {
	t.Parallel()

	c := &Client{Runner: &mockRunner{output: []byte("detached\n")}}
	if err := c.Detach(context.Background(), "00"); err != nil {
		t.Fatalf("Detach() error = %v", err)
	}
}

func TestClientDetachEmptyPort(t *testing.T) {
	t.Parallel()

	c := &Client{Runner: &mockRunner{}}
	if err := c.Detach(context.Background(), ""); err == nil {
		t.Fatal("Detach() expected error for empty port")
	}
}

func TestParseDevicePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		output string
		want   string
	}{
		{"ttyUSB0 path", "attached\n/dev/ttyUSB0\n", "/dev/ttyUSB0"},
		{"ttyACM0 path", "  /dev/ttyACM0  \n", "/dev/ttyACM0"},
		{"no device path", "attached to port 00\n", ""},
		{"empty output", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseDevicePath(tt.output)
			if got != tt.want {
				t.Errorf("parseDevicePath() = %q, want %q", got, tt.want)
			}
		})
	}
}
