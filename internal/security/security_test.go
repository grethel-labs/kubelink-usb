package security

import (
	"crypto/tls"
	"testing"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
)

func TestWhitelistAddAndHas(t *testing.T) {
	t.Parallel()

	w := NewWhitelist()
	if w.Has("device-a") {
		t.Fatal("expected device-a to be absent before Add")
	}

	w.Add("device-a")
	if !w.Has("device-a") {
		t.Fatal("expected device-a to be present after Add")
	}
}

func TestTLSConfigMinVersionTLS13(t *testing.T) {
	t.Parallel()

	cfg := TLSConfig()
	if cfg == nil {
		t.Fatal("expected non-nil TLS config")
	}
	if cfg.MinVersion != tls.VersionTLS13 {
		t.Fatalf("TLSConfig().MinVersion=%d want %d", cfg.MinVersion, tls.VersionTLS13)
	}
}

func TestEngineAllowsCurrentDefaultBehavior(t *testing.T) {
	t.Parallel()

	engine := &Engine{}
	if !engine.Allows(usbv1alpha1.USBDevice{}, usbv1alpha1.USBDevicePolicy{}) {
		t.Fatal("expected current Engine.Allows default to permit")
	}
}
