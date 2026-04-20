// Package security provides policy evaluation, whitelist management,
// and encryption helpers for USB device access control.
package security

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

// TLSConfig returns baseline secure TLS options for tunnel encryption.
func TLSConfig() *tls.Config {
	return &tls.Config{MinVersion: tls.VersionTLS13}
}

// MutualTLSConfig returns TLS 1.3 server config with required client cert validation.
func MutualTLSConfig(cert tls.Certificate, caPEM []byte) (*tls.Config, error) {
	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(caPEM); !ok {
		return nil, fmt.Errorf("invalid CA bundle")
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
	}, nil
}

// ClientTLSConfig returns TLS 1.3 client config with server CA validation.
func ClientTLSConfig(cert tls.Certificate, caPEM []byte, serverName string) (*tls.Config, error) {
	caPool := x509.NewCertPool()
	if ok := caPool.AppendCertsFromPEM(caPEM); !ok {
		return nil, fmt.Errorf("invalid CA bundle")
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
		ServerName:   serverName,
	}, nil
}
