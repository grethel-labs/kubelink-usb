package agent

import "context"

// Server manages usbipd export lifecycle on a node.
type Server struct{}

// Export publishes a local USB device through USB/IP.
//
// Intent: Encapsulate usbipd bind/export execution details.
// Inputs: Context and local host device path.
// Outputs: No value.
// Errors: Returns nil in the current stub behavior.
func (s *Server) Export(context.Context, string) error { return nil }

// Unexport removes a previously exported USB device from usbipd.
//
// Intent: Encapsulate export cleanup behavior.
// Inputs: Context and local host device path.
// Outputs: No value.
// Errors: Returns nil in the current stub behavior.
func (s *Server) Unexport(context.Context, string) error { return nil }
