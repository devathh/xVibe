package xvibe

import (
	"fmt"

	"github.com/devathh/xvibe/chat-gateway/internal/infrastructure/config"
	"github.com/devathh/xvibe/chat-gateway/internal/infrastructure/session/mtls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func ConnectInsecureXvibeChat(cfg *config.Config) (*grpc.ClientConn, error) {
	client, err := grpc.NewClient(
		fmt.Sprintf("%s:%d",
			cfg.Services.XvibeChat.Host,
			cfg.Services.XvibeChat.Port,
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create xvibe-chat client: %w", err)
	}

	return client, nil
}

func ConnectMTLSXvibeChat(cfg *config.Config) (*grpc.ClientConn, error) {
	tlsConfig, err := mtls.LoadMTLSConfig(cfg)
	if err != nil {
		return nil, err
	}

	client, err := grpc.NewClient(
		fmt.Sprintf("%s:%d",
			cfg.Services.XvibeChat.Host,
			cfg.Services.XvibeChat.Port,
		),
		grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create xvibe-chat client: %w", err)
	}

	return client, nil
}

func Close(cc *grpc.ClientConn) error {
	return cc.Close()
}
