package xvibe

import (
	"fmt"

	"github.com/devathh/xvibe/chat/internal/infrastructure/config"
	"github.com/devathh/xvibe/chat/internal/infrastructure/security/mtls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func ConnectInsecureAuth(cfg *config.Config) (*grpc.ClientConn, error) {
	clientConn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d",
			cfg.Services.XvibeAuth.Host,
			cfg.Services.XvibeAuth.Port,
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client conn: %w", err)
	}

	return clientConn, nil
}

func ConnectMTLSAuth(cfg *config.Config) (*grpc.ClientConn, error) {
	tlsCfg, err := mtls.LoadMTLSClient(cfg)
	if err != nil {
		return nil, err
	}

	clientConn, err := grpc.NewClient(
		fmt.Sprintf("%s:%d",
			cfg.Services.XvibeAuth.Host,
			cfg.Services.XvibeAuth.Port,
		),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client conn: %w", err)
	}

	return clientConn, nil
}

func Close(clientConn *grpc.ClientConn) error {
	if err := clientConn.Close(); err != nil {
		return fmt.Errorf("failed to close client conn: %w", err)
	}

	return nil
}
