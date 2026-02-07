package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/devathh/xvibe/auth-service/internal/infrastructure/config"
)

func LoadMTLSConfig(cfg *config.Config) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(
		cfg.Server.GRPC.TLS.ServerCert,
		cfg.Server.GRPC.TLS.ServerKey,
	)
	if err != nil {
		return nil, err
	}

	caCertPem, err := os.ReadFile(cfg.Server.GRPC.TLS.CaCert)
	if err != nil {
		return nil, err
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCertPem) {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
	}, nil
}
