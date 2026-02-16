package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"github.com/devathh/xvibe/chat/internal/infrastructure/config"
)

func LoadMTLSConfig(cfg *config.Config) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(
		cfg.Server.GRPC.TLS.ServerCert,
		cfg.Server.GRPC.TLS.ServerKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load server cert: %w", err)
	}

	caCertPem, err := os.ReadFile(cfg.Server.GRPC.TLS.CaCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read ca cert: %w", err)
	}

	return createCfg(caCertPem, cert)
}

func LoadMTLSClient(cfg *config.Config) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(
		cfg.Services.XvibeAuth.TLS.ClientCert,
		cfg.Services.XvibeAuth.TLS.ClientKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load server cert: %w", err)
	}

	caCertPem, err := os.ReadFile(cfg.Services.XvibeAuth.TLS.CaCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read ca cert: %w", err)
	}

	return createCfg(caCertPem, cert)
}

func createCfg(caCertPem []byte, cert tls.Certificate) (*tls.Config, error) {
	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCertPem) {
		return nil, errors.New("failed to append ca cert into pool")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}, nil
}
