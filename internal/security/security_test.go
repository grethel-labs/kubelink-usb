package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	usbv1alpha1 "github.com/grethel-labs/kubelink-usb/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestEngineAllows(t *testing.T) {
	t.Parallel()

	engine := &Engine{}

	tests := []struct {
		name   string
		device usbv1alpha1.USBDevice
		policy usbv1alpha1.USBDevicePolicy
		want   bool
	}{
		{
			name: "empty policy allows any device",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{VendorID: "04b4", ProductID: "6001", NodeName: "node-a"},
			},
			policy: usbv1alpha1.USBDevicePolicy{},
			want:   true,
		},
		{
			name: "vendor ID match",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{VendorID: "04b4", ProductID: "6001"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Selector: usbv1alpha1.USBDevicePolicySelector{VendorID: "04b4"},
				},
			},
			want: true,
		},
		{
			name: "vendor ID mismatch",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{VendorID: "1234"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Selector: usbv1alpha1.USBDevicePolicySelector{VendorID: "04b4"},
				},
			},
			want: false,
		},
		{
			name: "product ID mismatch",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{VendorID: "04b4", ProductID: "9999"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Selector: usbv1alpha1.USBDevicePolicySelector{VendorID: "04b4", ProductID: "6001"},
				},
			},
			want: false,
		},
		{
			name: "device on allowed node",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{NodeName: "node-a"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Restrictions: usbv1alpha1.USBDeviceRestrictions{AllowedNodes: []string{"node-a", "node-b"}},
				},
			},
			want: true,
		},
		{
			name: "device on denied node",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{NodeName: "node-c"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Restrictions: usbv1alpha1.USBDeviceRestrictions{AllowedNodes: []string{"node-a", "node-b"}},
				},
			},
			want: false,
		},
		{
			name: "node selector mismatch",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{NodeName: "node-c"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Selector: usbv1alpha1.USBDevicePolicySelector{NodeNames: []string{"node-a"}},
				},
			},
			want: false,
		},
		{
			name: "HID device denied",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{DeviceClass: "HID"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Restrictions: usbv1alpha1.USBDeviceRestrictions{DenyHumanInterfaceDevices: true},
				},
			},
			want: false,
		},
		{
			name: "HID device class 03 denied",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{DeviceClass: "03"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Restrictions: usbv1alpha1.USBDeviceRestrictions{DenyHumanInterfaceDevices: true},
				},
			},
			want: false,
		},
		{
			name: "non-HID device passes HID check",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{DeviceClass: "CDC"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Restrictions: usbv1alpha1.USBDeviceRestrictions{DenyHumanInterfaceDevices: true},
				},
			},
			want: true,
		},
		{
			name: "allowed device class",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{DeviceClass: "CDC"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Restrictions: usbv1alpha1.USBDeviceRestrictions{AllowedDeviceClasses: []string{"CDC", "vendor-specific"}},
				},
			},
			want: true,
		},
		{
			name: "disallowed device class",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{DeviceClass: "mass-storage"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Restrictions: usbv1alpha1.USBDeviceRestrictions{AllowedDeviceClasses: []string{"CDC"}},
				},
			},
			want: false,
		},
		{
			name: "case-insensitive vendor match",
			device: usbv1alpha1.USBDevice{
				Spec: usbv1alpha1.USBDeviceSpec{VendorID: "04B4"},
			},
			policy: usbv1alpha1.USBDevicePolicy{
				Spec: usbv1alpha1.USBDevicePolicySpec{
					Selector: usbv1alpha1.USBDevicePolicySelector{VendorID: "04b4"},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := engine.Allows(tt.device, tt.policy)
			if got != tt.want {
				t.Errorf("Allows() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEngineMatchesSelector(t *testing.T) {
	t.Parallel()

	engine := &Engine{}

	device := usbv1alpha1.USBDevice{
		ObjectMeta: metav1.ObjectMeta{Name: "dev1"},
		Spec:       usbv1alpha1.USBDeviceSpec{VendorID: "04b4", ProductID: "6001", NodeName: "node-a"},
	}

	matchingPolicy := usbv1alpha1.USBDevicePolicy{
		Spec: usbv1alpha1.USBDevicePolicySpec{
			Selector: usbv1alpha1.USBDevicePolicySelector{VendorID: "04b4"},
		},
	}
	nonMatchingPolicy := usbv1alpha1.USBDevicePolicy{
		Spec: usbv1alpha1.USBDevicePolicySpec{
			Selector: usbv1alpha1.USBDevicePolicySelector{VendorID: "dead"},
		},
	}

	if !engine.MatchesSelector(device, matchingPolicy) {
		t.Error("expected policy to match device")
	}
	if engine.MatchesSelector(device, nonMatchingPolicy) {
		t.Error("expected policy NOT to match device")
	}
}

func TestTLSHelpers(t *testing.T) {
	t.Parallel()

	cert, caPEM := generateTestCertAndCA(t)

	tests := []struct {
		name    string
		build   func() (*tls.Config, error)
		wantErr bool
	}{
		{
			name: "mutual tls config",
			build: func() (*tls.Config, error) {
				return MutualTLSConfig(cert, caPEM)
			},
		},
		{
			name: "client tls config",
			build: func() (*tls.Config, error) {
				return ClientTLSConfig(cert, caPEM, "kubelink-usb.local")
			},
		},
		{
			name: "invalid ca",
			build: func() (*tls.Config, error) {
				return MutualTLSConfig(cert, []byte("invalid"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg, err := tt.build()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.MinVersion != tls.VersionTLS13 {
				t.Fatalf("MinVersion=%d want=%d", cfg.MinVersion, tls.VersionTLS13)
			}
		})
	}
}

func TestBuildIsolationNetworkPolicy(t *testing.T) {
	t.Parallel()

	pol := BuildIsolationNetworkPolicy("default", "np-usb", map[string]string{"app": "kubelink-usb-agent"})
	if pol.Namespace != "default" || pol.Name != "np-usb" {
		t.Fatalf("unexpected metadata: %s/%s", pol.Namespace, pol.Name)
	}
	if len(pol.Spec.PolicyTypes) != 2 {
		t.Fatalf("policy types len=%d want=2", len(pol.Spec.PolicyTypes))
	}
	if got := pol.Spec.PodSelector.MatchLabels["app"]; got != "kubelink-usb-agent" {
		t.Fatalf("pod selector app=%q want kubelink-usb-agent", got)
	}
}

func generateTestCertAndCA(t *testing.T) (tls.Certificate, []byte) {
	t.Helper()

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey(ca) error = %v", err)
	}
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "kubelink-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("CreateCertificate(ca) error = %v", err)
	}

	certKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("GenerateKey(cert) error = %v", err)
	}
	certTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "kubelink-client"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, certTemplate, caTemplate, &certKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("CreateCertificate(cert) error = %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(certKey)})
	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf("X509KeyPair() error = %v", err)
	}
	return cert, caPEM
}
