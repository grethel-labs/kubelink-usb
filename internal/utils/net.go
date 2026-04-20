package utils

import "net"

// IsTCPReachable checks whether a TCP host:port can be dialed.
// Used by the health monitor to verify USB/IP export daemon availability.
func IsTCPReachable(addr string) bool {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
