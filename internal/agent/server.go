package agent

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// CommandRunner abstracts os/exec for testability.
type CommandRunner interface {
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
}

// ExecRunner executes real OS commands.
type ExecRunner struct{}

// Run executes a command and returns combined output.
func (r *ExecRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.Bytes(), err
}

// Server manages usbipd export lifecycle on a node.
// It wraps the usbipd CLI tool via the CommandRunner interface, executing
// bind/unbind operations to share or unshare local USB devices.
//
// @component AgentServer["Agent Export"] --> USBIPd["usbipd bind/unbind"]
type Server struct {
	Runner CommandRunner
}

// Export publishes a local USB device through USB/IP by binding it with usbipd.
//
// Intent: Encapsulate usbipd bind execution details.
// Inputs: Context and bus ID of the device to export.
// Outputs: No value.
// Errors: Returns command execution errors.
func (s *Server) Export(ctx context.Context, busID string) error {
	if busID == "" {
		return fmt.Errorf("busID must not be empty")
	}
	runner := s.runner()
	output, err := runner.Run(ctx, "usbipd", "bind", "--busid", busID)
	if err != nil {
		return fmt.Errorf("usbipd bind --busid %s: %w (output: %s)", busID, err, string(output))
	}
	return nil
}

// Unexport removes a previously exported USB device from usbipd.
//
// Intent: Encapsulate export cleanup behavior.
// Inputs: Context and bus ID of the device to unexport.
// Outputs: No value.
// Errors: Returns command execution errors.
func (s *Server) Unexport(ctx context.Context, busID string) error {
	if busID == "" {
		return fmt.Errorf("busID must not be empty")
	}
	runner := s.runner()
	output, err := runner.Run(ctx, "usbipd", "unbind", "--busid", busID)
	if err != nil {
		return fmt.Errorf("usbipd unbind --busid %s: %w (output: %s)", busID, err, string(output))
	}
	return nil
}

func (s *Server) runner() CommandRunner {
	if s.Runner != nil {
		return s.Runner
	}
	return &ExecRunner{}
}
