package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"github.com/devathh/xvibe/message-service/internal/infrastructure/config"
)

func LoadMTLSConfig(cfg *config.Config) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(
		cfg.Server.GRPC.TLS.ServerCert,
		cfg.Server.GRPC.TLS.ServerKey,
	)
	if err != nil {
		return nil, err
	}

	caCert, err := os.ReadFile(cfg.Server.GRPC.TLS.CaCert)
	if err != nil {
		return nil, err
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("failed to append ca cert into pool")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
	}, nil
}
