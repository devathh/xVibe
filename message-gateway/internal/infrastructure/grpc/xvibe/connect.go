package xvibegrpc

import (
	"fmt"

	"github.com/devathh/xvibe/message-gateway/internal/infrastructure/config"
	"github.com/devathh/xvibe/message-gateway/internal/infrastructure/security/mtls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func ConnectInsecureMessage(cfg *config.Config) (*grpc.ClientConn, error) {
	client, err := grpc.NewClient(
		fmt.Sprintf("%s:%d",
			cfg.Services.XVBMessage.Host,
			cfg.Services.XVBMessage.Port,
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client conn: %w", err)
	}

	return client, nil
}

func ConnectMTLSMessage(cfg *config.Config) (*grpc.ClientConn, error) {
	tlsConfig, err := mtls.LoadMTLSConfig(cfg)
	if err != nil {
		return nil, err
	}

	client, err := grpc.NewClient(
		fmt.Sprintf("%s:%d",
			cfg.Services.XVBMessage.Host,
			cfg.Services.XVBMessage.Port,
		),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client conn: %w", err)
	}

	return client, nil
}

func Close(clientConn *grpc.ClientConn) error {
	if err := clientConn.Close(); err != nil {
		return fmt.Errorf("failed to close client conn: %w", err)
	}

	return nil
}
