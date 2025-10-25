package certs

import (
	"crypto/x509"
	"fmt"
	"time"
)

// ValidateCertificate validates a certificate against a CA pool
func ValidateCertificate(cert *x509.Certificate, caPool *x509.CertPool) error {
	opts := x509.VerifyOptions{
		Roots:     caPool,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	if _, err := cert.Verify(opts); err != nil {
		return fmt.Errorf("certificate validation failed: %w", err)
	}

	return nil
}

// ValidateCertificateExpiry checks if a certificate is expired or about to expire
func ValidateCertificateExpiry(cert *x509.Certificate, warningThreshold time.Duration) error {
	now := time.Now()

	if now.Before(cert.NotBefore) {
		return fmt.Errorf("certificate not yet valid (valid from %v)", cert.NotBefore)
	}

	if now.After(cert.NotAfter) {
		return fmt.Errorf("certificate expired on %v", cert.NotAfter)
	}

	// Check if certificate expires soon
	timeUntilExpiry := cert.NotAfter.Sub(now)
	if timeUntilExpiry < warningThreshold {
		return fmt.Errorf("certificate expires in %v (threshold: %v)", timeUntilExpiry, warningThreshold)
	}

	return nil
}

// ExtractCommonName extracts the Common Name from a certificate
func ExtractCommonName(cert *x509.Certificate) string {
	return cert.Subject.CommonName
}

// ExtractOrganization extracts the Organization from a certificate
func ExtractOrganization(cert *x509.Certificate) string {
	if len(cert.Subject.Organization) > 0 {
		return cert.Subject.Organization[0]
	}
	return ""
}

// IsCACertificate checks if a certificate is a CA certificate
func IsCACertificate(cert *x509.Certificate) bool {
	return cert.IsCA
}

// GetCertificateInfo returns formatted information about a certificate
func GetCertificateInfo(cert *x509.Certificate) map[string]interface{} {
	return map[string]interface{}{
		"subject":      cert.Subject.String(),
		"issuer":       cert.Issuer.String(),
		"serial":       cert.SerialNumber.String(),
		"not_before":   cert.NotBefore,
		"not_after":    cert.NotAfter,
		"is_ca":        cert.IsCA,
		"dns_names":    cert.DNSNames,
		"organization": ExtractOrganization(cert),
		"common_name":  ExtractCommonName(cert),
	}
}