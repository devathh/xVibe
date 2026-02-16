package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/devathh/xvibe/chat-gateway/internal/infrastructure/config"
)

func LoadMTLSConfig(cfg *config.Config) (*tls.Config, error) {
	clientCert, err := tls.LoadX509KeyPair(
		cfg.Services.XvibeChat.TLS.ClientCert,
		cfg.Services.XvibeChat.TLS.ClientKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert: %w", err)
	}

	caPem, err := os.ReadFile(cfg.Services.XvibeChat.TLS.CaCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read ca cert: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caPem) {
		return nil, fmt.Errorf("failed to append cert into pool: %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		ClientCAs:    caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}, nil
}
