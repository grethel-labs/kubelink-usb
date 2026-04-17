package agent

import "context"

// Client manages vhci attachment lifecycle on a node.
type Client struct{}

func (c *Client) Attach(context.Context, string, string) (string, error) { return "", nil }
func (c *Client) Detach(context.Context, string) error                   { return nil }
