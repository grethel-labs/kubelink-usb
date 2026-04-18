package usbip

import "context"

// Serve starts the USB/IP export serving loop.
//
// Intent: Provide a stable interface for export daemon lifecycle.
// Inputs: Context for cancellation and bind address.
// Outputs: Nil on the current scaffold implementation.
// Errors: Returns nil in the current stub behavior.
func Serve(context.Context, string) error { return nil }
