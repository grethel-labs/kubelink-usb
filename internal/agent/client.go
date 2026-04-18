package agent

import "context"

// Client manages vhci attachment lifecycle on a node.
type Client struct{}

// Attach imports a remote USB/IP bus and returns the mapped client device path.
//
// Intent: Centralize future vhci attach logic behind a small interface.
// Inputs: Context, server endpoint, and exported bus identifier.
// Outputs: Empty device path in the current stub behavior.
// Errors: Returns nil in the current stub behavior.
func (c *Client) Attach(context.Context, string, string) (string, error) { return "", nil }

// Detach removes a previously attached vhci port mapping.
//
// Intent: Centralize detach lifecycle and cleanup paths.
// Inputs: Context and client vhci port identifier.
// Outputs: No value.
// Errors: Returns nil in the current stub behavior.
func (c *Client) Detach(context.Context, string) error { return nil }
