package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/devathh/xvibe/message-gateway/internal/infrastructure/config"
)

type Server struct {
	cfg *config.Config
	srv *http.Server
}

func New(cfg *config.Config, handler http.Handler) *Server {
	return &Server{
		cfg: cfg,
		srv: &http.Server{
			Addr: fmt.Sprintf("%s:%d",
				cfg.Server.HTTP.Host,
				cfg.Server.HTTP.Port,
			),
			Handler:      handler,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
		},
	}
}

func (s *Server) Start() error {
	if s.cfg.Server.HTTP.TLS.Enable {
		return s.srv.ListenAndServeTLS(
			s.cfg.Server.HTTP.TLS.ServerCert,
			s.cfg.Server.HTTP.TLS.ServerKey,
		)
	}

	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
