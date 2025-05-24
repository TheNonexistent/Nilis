package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	cfg "github.com/thenonexistent/nilis/internal/config"
	"google.golang.org/grpc/credentials"
)

func newServerTLS(config *cfg.Config) (credentials.TransportCredentials, error) {
	serverCert, err := tls.LoadX509KeyPair(config.Server.TLSCert, config.Server.TLSKey)
	if err != nil {
		return nil, fmt.Errorf("failed initializing server certificate: %w", err)
	}

	caCert, err := os.ReadFile(config.Server.TLSCA)
	if err != nil {
		return nil, fmt.Errorf("failed loading ca certificate: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("failed appending ca to x509 certificate pool")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		RootCAs:      certPool,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS13,
		MaxVersion:   tls.VersionTLS13,
	}

	return credentials.NewTLS(tlsConfig), nil
}
