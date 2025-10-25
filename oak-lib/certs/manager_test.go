package certs

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Verify CA cert is generated
	caCert := manager.GetCACert()
	if len(caCert) == 0 {
		t.Error("CA cert should not be empty")
	}

	// Verify server cert is generated
	serverCert, serverKey := manager.GetServerCertAndKey()
	if len(serverCert) == 0 {
		t.Error("Server cert should not be empty")
	}
	if len(serverKey) == 0 {
		t.Error("Server key should not be empty")
	}

	// Verify certs are valid PEM
	if !isPEMEncoded(caCert) {
		t.Error("CA cert should be PEM encoded")
	}
	if !isPEMEncoded(serverCert) {
		t.Error("Server cert should be PEM encoded")
	}
	if !isPEMEncoded(serverKey) {
		t.Error("Server key should be PEM encoded")
	}
}

func TestGenerateClientCert(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	tests := []struct {
		name    string
		agentID string
		wantErr bool
	}{
		{
			name:    "valid client cert",
			agentID: "agent-001",
			wantErr: false,
		},
		{
			name:    "another valid client cert",
			agentID: "agent-002",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientCert, clientKey, err := manager.GenerateClientCert(tt.agentID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateClientCert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Verify cert and key are not empty
			if len(clientCert) == 0 {
				t.Error("Client cert should not be empty")
			}
			if len(clientKey) == 0 {
				t.Error("Client key should not be empty")
			}

			// Verify PEM encoding
			if !isPEMEncoded(clientCert) {
				t.Error("Client cert should be PEM encoded")
			}
			if !isPEMEncoded(clientKey) {
				t.Error("Client key should be PEM encoded")
			}

			// Parse and verify certificate
			block, _ := pem.Decode(clientCert)
			if block == nil {
				t.Fatal("Failed to decode PEM block")
			}

			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				t.Fatalf("Failed to parse certificate: %v", err)
			}

			// Verify certificate fields (manager prefixes with "agent-")
			expectedCN := fmt.Sprintf("agent-%s", tt.agentID)
			if cert.Subject.CommonName != expectedCN {
				t.Errorf("CommonName = %s, want %s", cert.Subject.CommonName, expectedCN)
			}

			// Verify expiration (should be ~90 days)
			duration := cert.NotAfter.Sub(cert.NotBefore)
			expectedDuration := 90 * 24 * time.Hour
			if duration < expectedDuration-time.Hour || duration > expectedDuration+time.Hour {
				t.Errorf("Certificate duration = %v, want ~%v", duration, expectedDuration)
			}

			// Verify it's not expired
			if time.Now().After(cert.NotAfter) {
				t.Error("Certificate is expired")
			}
			if time.Now().Before(cert.NotBefore) {
				t.Error("Certificate is not yet valid")
			}
		})
	}
}

func TestClientCertVerification(t *testing.T) {
	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Generate client cert
	clientCertPEM, _, err := manager.GenerateClientCert("test-agent")
	if err != nil {
		t.Fatalf("GenerateClientCert() error = %v", err)
	}

	// Parse CA cert
	caCertPEM := manager.GetCACert()
	caBlock, _ := pem.Decode(caCertPEM)
	if caBlock == nil {
		t.Fatal("Failed to decode CA PEM")
	}

	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse CA cert: %v", err)
	}

	// Parse client cert
	clientBlock, _ := pem.Decode(clientCertPEM)
	if clientBlock == nil {
		t.Fatal("Failed to decode client PEM")
	}

	clientCert, err := x509.ParseCertificate(clientBlock.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse client cert: %v", err)
	}

	// Verify client cert was signed by CA
	roots := x509.NewCertPool()
	roots.AddCert(caCert)

	opts := x509.VerifyOptions{
		Roots:     roots,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	if _, err := clientCert.Verify(opts); err != nil {
		t.Errorf("Client cert verification failed: %v", err)
	}
}

func TestConcurrentClientCertGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	manager, err := NewManager()
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Generate multiple certs concurrently
	const numCerts = 10
	errChan := make(chan error, numCerts)

	for i := 0; i < numCerts; i++ {
		go func(id int) {
			agentID := string(rune('a' + id))
			_, _, err := manager.GenerateClientCert(agentID)
			errChan <- err
		}(i)
	}

	// Check all succeeded
	for i := 0; i < numCerts; i++ {
		if err := <-errChan; err != nil {
			t.Errorf("Concurrent cert generation failed: %v", err)
		}
	}
}

// Helper function to check if data is PEM encoded
func isPEMEncoded(data []byte) bool {
	block, _ := pem.Decode(data)
	return block != nil
}