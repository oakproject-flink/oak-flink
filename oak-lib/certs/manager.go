package certs

import (
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"
	"sync"
	"time"
)

// Manager manages CA and server certificates
type Manager struct {
	mu sync.RWMutex

	caCert     *x509.Certificate
	caKey      *ecdsa.PrivateKey
	serverCert *x509.Certificate
	serverKey  *ecdsa.PrivateKey

	caCertPEM     []byte
	serverCertPEM []byte
	serverKeyPEM  []byte
}

// NewManager creates a new certificate manager
func NewManager() (*Manager, error) {
	m := &Manager{}

	// Generate CA
	caCert, caKey, err := GenerateCA(CAConfig{
		Organization: "Oak Platform",
		CommonName:   "Oak Root CA",
		ValidFor:     10 * 365 * 24 * time.Hour, // 10 years
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate CA: %w", err)
	}

	m.caCert = caCert
	m.caKey = caKey
	m.caCertPEM = EncodeCertToPEM(caCert)

	// Generate server certificate
	serverCert, serverKey, err := GenerateServerCert(
		CertConfig{
			Organization: "Oak Platform",
			CommonName:   "oak-server",
			DNSNames:     []string{"oak-server", "oak-server.oak-system", "oak-server.oak-system.svc.cluster.local", "localhost"},
			ValidFor:     365 * 24 * time.Hour, // 1 year
		},
		caCert,
		caKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate server certificate: %w", err)
	}

	m.serverCert = serverCert
	m.serverKey = serverKey
	m.serverCertPEM = EncodeCertToPEM(serverCert)
	serverKeyPEM, err := EncodePrivateKeyToPEM(serverKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode server key: %w", err)
	}
	m.serverKeyPEM = serverKeyPEM

	return m, nil
}

// GenerateClientCert generates a new client certificate for an agent
func (m *Manager) GenerateClientCert(agentID string) (certPEM, keyPEM []byte, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cert, key, err := GenerateClientCert(
		CertConfig{
			Organization: "Oak Platform",
			CommonName:   fmt.Sprintf("agent-%s", agentID),
			ValidFor:     90 * 24 * time.Hour, // 90 days
		},
		m.caCert,
		m.caKey,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate client certificate: %w", err)
	}

	certPEM = EncodeCertToPEM(cert)
	keyPEM, err = EncodePrivateKeyToPEM(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode client key: %w", err)
	}

	return certPEM, keyPEM, nil
}

// GetCACert returns the CA certificate PEM
func (m *Manager) GetCACert() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.caCertPEM
}

// GetServerCert returns the server certificate PEM
func (m *Manager) GetServerCert() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.serverCertPEM
}

// GetServerKey returns the server private key PEM
func (m *Manager) GetServerKey() []byte {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.serverKeyPEM
}

// GetServerCertAndKey returns both server cert and key PEM
func (m *Manager) GetServerCertAndKey() ([]byte, []byte) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.serverCertPEM, m.serverKeyPEM
}
