package agent

import (
	"context"
	"fmt"
	"strings"
)

// Client manages vhci attachment lifecycle on a node.
type Client struct {
	Runner CommandRunner
}

// Attach imports a remote USB/IP bus and returns the mapped client device path.
//
// Intent: Centralize vhci attach logic and device path parsing.
// Inputs: Context, server endpoint (host:port), and exported bus identifier.
// Outputs: Mapped device path (e.g., "/dev/ttyUSB0") parsed from usbip output.
// Errors: Returns command execution or parsing errors.
func (c *Client) Attach(ctx context.Context, remote string, busID string) (string, error) {
	if remote == "" {
		return "", fmt.Errorf("remote endpoint must not be empty")
	}
	if busID == "" {
		return "", fmt.Errorf("busID must not be empty")
	}

	runner := c.runner()
	output, err := runner.Run(ctx, "usbip", "attach", "--remote", remote, "--busid", busID)
	if err != nil {
		return "", fmt.Errorf("usbip attach --remote %s --busid %s: %w (output: %s)", remote, busID, err, string(output))
	}

	devicePath := parseDevicePath(string(output))
	return devicePath, nil
}

// Detach removes a previously attached vhci port mapping.
//
// Intent: Centralize detach lifecycle and cleanup paths.
// Inputs: Context and vhci port identifier.
// Outputs: No value.
// Errors: Returns command execution errors.
func (c *Client) Detach(ctx context.Context, port string) error {
	if port == "" {
		return fmt.Errorf("port must not be empty")
	}

	runner := c.runner()
	output, err := runner.Run(ctx, "usbip", "detach", "--port", port)
	if err != nil {
		return fmt.Errorf("usbip detach --port %s: %w (output: %s)", port, err, string(output))
	}
	return nil
}

func (c *Client) runner() CommandRunner {
	if c.Runner != nil {
		return c.Runner
	}
	return &ExecRunner{}
}

// parseDevicePath extracts a /dev/ttyUSB* or /dev/ttyACM* path from usbip output.
func parseDevicePath(output string) string {
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "/dev/ttyUSB") || strings.HasPrefix(trimmed, "/dev/ttyACM") {
			return trimmed
		}
	}
	return ""
}
