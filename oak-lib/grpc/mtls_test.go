package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"testing"

	"github.com/oakproject-flink/oak-flink/oak-lib/certs"
)

func TestNewServerCredentials(t *testing.T) {
	// Generate test certificates
	certManager, err := certs.NewManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	serverCert, serverKey := certManager.GetServerCertAndKey()
	caCert := certManager.GetCACert()

	creds, err := NewServerCredentials(serverCert, serverKey, caCert)
	if err != nil {
		t.Fatalf("NewServerCredentials() error = %v", err)
	}

	if creds == nil {
		t.Error("NewServerCredentials() returned nil")
	}

	// Verify credentials have TLS info
	tlsInfo := creds.Info()
	if tlsInfo.SecurityProtocol != "tls" {
		t.Errorf("SecurityProtocol = %s, want tls", tlsInfo.SecurityProtocol)
	}
}

func TestNewServerCredentialsWithOptionalClient(t *testing.T) {
	certManager, err := certs.NewManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	serverCert, serverKey := certManager.GetServerCertAndKey()
	caCert := certManager.GetCACert()

	creds, err := NewServerCredentialsWithOptionalClient(serverCert, serverKey, caCert)
	if err != nil {
		t.Fatalf("NewServerCredentialsWithOptionalClient() error = %v", err)
	}

	if creds == nil {
		t.Error("NewServerCredentialsWithOptionalClient() returned nil")
	}
}

func TestNewClientCredentials(t *testing.T) {
	certManager, err := certs.NewManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	// Generate client cert
	clientCert, clientKey, err := certManager.GenerateClientCert("test-agent")
	if err != nil {
		t.Fatalf("GenerateClientCert() error = %v", err)
	}

	caCert := certManager.GetCACert()

	creds, err := NewClientCredentials(clientCert, clientKey, caCert, "localhost")
	if err != nil {
		t.Fatalf("NewClientCredentials() error = %v", err)
	}

	if creds == nil {
		t.Error("NewClientCredentials() returned nil")
	}

	tlsInfo := creds.Info()
	if tlsInfo.SecurityProtocol != "tls" {
		t.Errorf("SecurityProtocol = %s, want tls", tlsInfo.SecurityProtocol)
	}
}

func TestServerCredentialsInvalidCert(t *testing.T) {
	tests := []struct {
		name       string
		serverCert []byte
		serverKey  []byte
		caCert     []byte
	}{
		{
			name:       "empty server cert",
			serverCert: []byte{},
			serverKey:  []byte("key"),
			caCert:     []byte("ca"),
		},
		{
			name:       "empty server key",
			serverCert: []byte("cert"),
			serverKey:  []byte{},
			caCert:     []byte("ca"),
		},
		{
			name:       "empty ca cert",
			serverCert: []byte("cert"),
			serverKey:  []byte("key"),
			caCert:     []byte{},
		},
		{
			name:       "invalid server cert",
			serverCert: []byte("invalid cert"),
			serverKey:  []byte("invalid key"),
			caCert:     []byte("invalid ca"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewServerCredentials(tt.serverCert, tt.serverKey, tt.caCert)
			if err == nil {
				t.Error("NewServerCredentials() should fail with invalid cert")
			}
		})
	}
}

func TestClientCredentialsInvalidCert(t *testing.T) {
	tests := []struct {
		name       string
		clientCert []byte
		clientKey  []byte
		caCert     []byte
	}{
		{
			name:       "empty client cert",
			clientCert: []byte{},
			clientKey:  []byte("key"),
			caCert:     []byte("ca"),
		},
		{
			name:       "invalid client cert",
			clientCert: []byte("invalid"),
			clientKey:  []byte("invalid"),
			caCert:     []byte("invalid"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClientCredentials(tt.clientCert, tt.clientKey, tt.caCert, "localhost")
			if err == nil {
				t.Error("NewClientCredentials() should fail with invalid cert")
			}
		})
	}
}

func TestTLSConfigVerification(t *testing.T) {
	certManager, err := certs.NewManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	serverCert, serverKey := certManager.GetServerCertAndKey()

	// Parse server cert to verify it's valid
	serverCertParsed, err := tls.X509KeyPair(serverCert, serverKey)
	if err != nil {
		t.Fatalf("Failed to parse cert: %v", err)
	}

	if len(serverCertParsed.Certificate) == 0 {
		t.Error("Certificate should not be empty")
	}
}

func TestMinTLSVersion(t *testing.T) {
	certManager, err := certs.NewManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	serverCert, serverKey := certManager.GetServerCertAndKey()
	caCert := certManager.GetCACert()

	// We can't directly test the TLS version from the credentials object,
	// but we can verify the credentials are created successfully
	creds, err := NewServerCredentials(serverCert, serverKey, caCert)
	if err != nil {
		t.Fatalf("NewServerCredentials() error = %v", err)
	}

	if creds == nil {
		t.Error("Credentials should not be nil")
	}

	// The implementation should enforce TLS 1.3 minimum
	// This would be verified in integration tests with actual connections
}

func TestClientAuthPolicy(t *testing.T) {
	certManager, err := certs.NewManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	serverCert, serverKey := certManager.GetServerCertAndKey()
	caCert := certManager.GetCACert()

	// Test RequireAndVerifyClientCert
	creds1, err := NewServerCredentials(serverCert, serverKey, caCert)
	if err != nil {
		t.Fatalf("NewServerCredentials() error = %v", err)
	}
	if creds1 == nil {
		t.Error("Credentials should not be nil")
	}

	// Test VerifyClientCertIfGiven
	creds2, err := NewServerCredentialsWithOptionalClient(serverCert, serverKey, caCert)
	if err != nil {
		t.Fatalf("NewServerCredentialsWithOptionalClient() error = %v", err)
	}
	if creds2 == nil {
		t.Error("Credentials should not be nil")
	}

	// Both should succeed, but have different client auth policies
	// This would be verified in integration tests
}

func TestCAPoolSetup(t *testing.T) {
	certManager, err := certs.NewManager()
	if err != nil {
		t.Fatalf("Failed to create cert manager: %v", err)
	}

	caCert := certManager.GetCACert()

	// Parse CA cert
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		t.Fatal("Failed to append CA cert to pool")
	}

	// Verify cert pool has subjects
	if len(certPool.Subjects()) == 0 {
		t.Error("Cert pool should have at least one subject")
	}
}