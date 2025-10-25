package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ServerTLSConfig creates a TLS configuration for gRPC server with mTLS
func ServerTLSConfig(serverCert, serverKey, caCert []byte) (*tls.Config, error) {
	// Load server certificate
	cert, err := tls.X509KeyPair(serverCert, serverKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %w", err)
	}

	// Load CA certificate for client verification
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to add CA certificate to pool")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS13,
	}, nil
}

// ClientTLSConfig creates a TLS configuration for gRPC client with mTLS
func ClientTLSConfig(clientCert, clientKey, caCert []byte, serverName string) (*tls.Config, error) {
	// Load client certificate
	cert, err := tls.X509KeyPair(clientCert, clientKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	// Load CA certificate for server verification
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to add CA certificate to pool")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
		ServerName:   serverName,
		MinVersion:   tls.VersionTLS13,
	}, nil
}

// NewServerCredentials creates gRPC server credentials with mTLS (requires client cert)
func NewServerCredentials(serverCert, serverKey, caCert []byte) (credentials.TransportCredentials, error) {
	tlsConfig, err := ServerTLSConfig(serverCert, serverKey, caCert)
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(tlsConfig), nil
}

// NewServerCredentialsWithOptionalClient creates gRPC server credentials with optional client cert
// This allows agents to connect without certs for initial registration
func NewServerCredentialsWithOptionalClient(serverCert, serverKey, caCert []byte) (credentials.TransportCredentials, error) {
	// Load server certificate
	cert, err := tls.X509KeyPair(serverCert, serverKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %w", err)
	}

	// Load CA certificate for client verification (when provided)
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to add CA certificate to pool")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.VerifyClientCertIfGiven, // Optional client cert
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS13,
	}

	return credentials.NewTLS(tlsConfig), nil
}

// NewClientCredentials creates gRPC client credentials with mTLS
func NewClientCredentials(clientCert, clientKey, caCert []byte, serverName string) (credentials.TransportCredentials, error) {
	tlsConfig, err := ClientTLSConfig(clientCert, clientKey, caCert, serverName)
	if err != nil {
		return nil, err
	}

	return credentials.NewTLS(tlsConfig), nil
}

// ServerOptions returns gRPC server options with mTLS
func ServerOptions(creds credentials.TransportCredentials) []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.MaxRecvMsgSize(10 * 1024 * 1024), // 10MB
		grpc.MaxSendMsgSize(10 * 1024 * 1024), // 10MB
	}
}

// ClientOptions returns gRPC client options with mTLS
func ClientOptions(creds credentials.TransportCredentials) []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(10*1024*1024), // 10MB
			grpc.MaxCallSendMsgSize(10*1024*1024), // 10MB
		),
	}
}
