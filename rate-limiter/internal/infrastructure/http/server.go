package httpserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/devathh/xvibe/rate-limiter/internal/infrastructure/config"
)

// The main server
type Server struct {
	isTLS                 bool
	serverCert, serverKey string
	srv                   *http.Server
}

func New(cfg *config.Config, handler http.Handler) *Server {
	return &Server{
		isTLS:      cfg.Server.HTTP.TLS.Enable,
		serverCert: cfg.Server.HTTP.TLS.ServerCert,
		serverKey:  cfg.Server.HTTP.TLS.ServerKey,
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
	if s.isTLS {
		return s.srv.ListenAndServeTLS(
			s.serverCert,
			s.serverKey,
		)
	}

	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
