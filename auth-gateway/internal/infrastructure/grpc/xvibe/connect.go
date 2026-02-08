package xvibegrpc

import (
	"fmt"
	"net"

	"github.com/devathh/xvibe/auth-gateway/internal/infrastructure/config"
	"github.com/devathh/xvibe/auth-gateway/internal/infrastructure/security/mtls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func ConnectInsecureAuth(cfg *config.Config) (*grpc.ClientConn, error) {
	client, err := grpc.NewClient(
		net.JoinHostPort(
			cfg.Services.XvibeAuth.Host,
			fmt.Sprint(cfg.Services.XvibeAuth.Port),
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client conn with xvibe-auth: %w", err)
	}

	return client, nil
}

func ConnectMTLSAuth(cfg *config.Config) (*grpc.ClientConn, error) {
	mtlsConfig, err := mtls.LoadMTLSConfig(cfg)
	if err != nil {
		return nil, err
	}

	client, err := grpc.NewClient(
		net.JoinHostPort(
			cfg.Services.XvibeAuth.Host,
			fmt.Sprint(cfg.Services.XvibeAuth.Port),
		),
		grpc.WithTransportCredentials(credentials.NewTLS(mtlsConfig)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client conn with xvibe-auth: %w", err)
	}

	return client, nil
}

func CloseAuth(conn *grpc.ClientConn) error {
	if err := conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection with xvibe-auth: %w", err)
	}

	return nil
}
