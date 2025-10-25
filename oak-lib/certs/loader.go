package certs

import (
	"crypto/ecdsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// LoadCertificateFromFile loads a certificate from a PEM file
func LoadCertificateFromFile(path string) (*x509.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	return LoadCertificateFromPEM(data)
}

// LoadCertificateFromPEM loads a certificate from PEM bytes
func LoadCertificateFromPEM(pemData []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("unexpected PEM block type: %s", block.Type)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// LoadPrivateKeyFromFile loads an ECDSA private key from a PEM file
func LoadPrivateKeyFromFile(path string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	return LoadPrivateKeyFromPEM(data)
}

// LoadPrivateKeyFromPEM loads an ECDSA private key from PEM bytes
func LoadPrivateKeyFromPEM(pemData []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	if block.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("unexpected PEM block type: %s", block.Type)
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EC private key: %w", err)
	}

	return key, nil
}

// LoadTLSCertificate loads a tls.Certificate from cert and key files
func LoadTLSCertificate(certPath, keyPath string) (tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	return cert, nil
}

// LoadTLSCertificateFromPEM loads a tls.Certificate from PEM bytes
func LoadTLSCertificateFromPEM(certPEM, keyPEM []byte) (tls.Certificate, error) {
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("failed to load TLS certificate from PEM: %w", err)
	}

	return cert, nil
}

// LoadCAPool loads a CA certificate pool from a file
func LoadCAPool(caPath string) (*x509.CertPool, error) {
	caCert, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	return LoadCAPoolFromPEM(caCert)
}

// LoadCAPoolFromPEM loads a CA certificate pool from PEM bytes
func LoadCAPoolFromPEM(caPEM []byte) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("failed to append CA certificate to pool")
	}

	return certPool, nil
}

// SaveCertificateToFile saves a certificate to a PEM file
func SaveCertificateToFile(cert *x509.Certificate, path string) error {
	certPEM := EncodeCertToPEM(cert)
	return os.WriteFile(path, certPEM, 0644)
}

// SavePrivateKeyToFile saves a private key to a PEM file
func SavePrivateKeyToFile(key *ecdsa.PrivateKey, path string) error {
	keyPEM, err := EncodePrivateKeyToPEM(key)
	if err != nil {
		return err
	}
	return os.WriteFile(path, keyPEM, 0600) // More restrictive permissions for private key
}