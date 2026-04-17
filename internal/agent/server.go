package agent

import "context"

// Server manages usbipd export lifecycle on a node.
type Server struct{}

func (s *Server) Export(context.Context, string) error   { return nil }
func (s *Server) Unexport(context.Context, string) error { return nil }
