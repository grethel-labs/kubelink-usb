package security

import "crypto/tls"

// TLSConfig returns baseline secure TLS options for tunnel encryption.
func TLSConfig() *tls.Config {
	return &tls.Config{MinVersion: tls.VersionTLS13}
}
