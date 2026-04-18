package utils

import "net"

// IsTCPReachable checks whether a TCP host:port can be dialed.
func IsTCPReachable(addr string) bool {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
