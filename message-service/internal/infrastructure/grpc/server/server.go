package grpcserver

import (
	"fmt"
	"net"

	"github.com/devathh/xvibe/message-service/internal/infrastructure/config"
	"google.golang.org/grpc"
)

type Server struct {
	cfg     *config.Config
	grpcSrv *grpc.Server
}

func New(cfg *config.Config, grpcSrv *grpc.Server) *Server {
	return &Server{
		cfg:     cfg,
		grpcSrv: grpcSrv,
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen(
		s.cfg.Server.GRPC.Protocol,
		fmt.Sprintf("%s:%d",
			s.cfg.Server.GRPC.Host,
			s.cfg.Server.GRPC.Port,
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	if err := s.grpcSrv.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve listener: %w", err)
	}

	return nil
}

func (s *Server) GracefulShutdown() {
	s.grpcSrv.GracefulStop()
}
