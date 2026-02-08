package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

var (
	// App's
	ErrAppInvalidName    = errors.New("invalid name of service")
	ErrAppInvalidVersion = errors.New("invalid version of service")

	// Server's
	ErrServerInvalidServerCert     = errors.New("invalid path to server cert")
	ErrServerInvalidServerKey      = errors.New("invalid path to server key")
	ErrServerTooLittleReadTimeout  = errors.New("too little read-timeout")
	ErrServerTooLittleWriteTimeout = errors.New("too little write-timeout")
	ErrServerTooLittleIdleTimeout  = errors.New("too little idle-timeout")

	// Service's
	ErrServiceInvalidHost       = errors.New("invalid host")
	ErrServiceInvalidPort       = errors.New("invalid port")
	ErrServiceInvalidClientCert = errors.New("invalid path to client cert")
	ErrServiceInvalidClientKey  = errors.New("invalid path to client key")
	ErrServiceInvalidCaCert     = errors.New("invalid path to ca cert")

	// General's
	ErrInvalidPath = errors.New("invalid path to config file")
)

type app struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func (a *app) validate() error {
	a.Name = strings.TrimSpace(a.Name)
	if a.Name == "" {
		return ErrAppInvalidName
	}

	a.Version = strings.TrimSpace(a.Version)
	if a.Version == "" {
		return ErrAppInvalidVersion
	}

	return nil
}

type server struct {
	HTTP struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
		TLS  struct {
			Enable     bool   `yaml:"enable"`
			ServerCert string `yaml:"server-cert"`
			ServerKey  string `yaml:"server-key"`
		} `yaml:"tls"`
	} `yaml:"http"`
	ReadTimeout  time.Duration `yaml:"read-timeout"`
	WriteTimeout time.Duration `yaml:"write-timeout"`
	IdleTimeout  time.Duration `yaml:"idle-timeout"`
}

func (s *server) applyDefaults() {
	s.HTTP.Host = strings.TrimSpace(s.HTTP.Host)
	if s.HTTP.Host == "" {
		s.HTTP.Host = "localhost"
	}

	if s.HTTP.Port <= 0 || s.HTTP.Port > 65535 {
		s.HTTP.Port = 7070
	}
}

func (s *server) validate() error {
	if s.HTTP.TLS.Enable {
		s.HTTP.TLS.ServerCert = strings.TrimSpace(s.HTTP.TLS.ServerCert)
		if s.HTTP.TLS.ServerCert == "" {
			return ErrServerInvalidServerCert
		}

		s.HTTP.TLS.ServerKey = strings.TrimSpace(s.HTTP.TLS.ServerKey)
		if s.HTTP.TLS.ServerKey == "" {
			return ErrServerInvalidServerKey
		}
	}

	if s.ReadTimeout < 100*time.Millisecond {
		return ErrServerTooLittleReadTimeout
	}

	if s.WriteTimeout < 100*time.Millisecond {
		return ErrServerTooLittleWriteTimeout
	}

	if s.IdleTimeout < 100*time.Millisecond {
		return ErrServerTooLittleIdleTimeout
	}

	return nil
}

type xvibeAuth struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	MTLS struct {
		Enable     bool   `yaml:"enable"`
		ClientCert string `yaml:"client-cert"`
		ClientKey  string `yaml:"client-key"`
		CaCert     string `yaml:"ca-cert"`
	} `yaml:"mtls"`
}

func (xa *xvibeAuth) validate() error {
	if xa.MTLS.Enable {
		xa.MTLS.ClientCert = strings.TrimSpace(xa.MTLS.ClientCert)
		if xa.MTLS.ClientCert == "" {
			return ErrServiceInvalidClientCert
		}

		xa.MTLS.ClientKey = strings.TrimSpace(xa.MTLS.ClientKey)
		if xa.MTLS.ClientKey == "" {
			return ErrServiceInvalidClientKey
		}

		xa.MTLS.CaCert = strings.TrimSpace(xa.MTLS.CaCert)
		if xa.MTLS.CaCert == "" {
			return ErrServiceInvalidCaCert
		}
	}
	xa.Host = strings.TrimSpace(xa.Host)
	if xa.Host == "" {
		return ErrServiceInvalidHost
	}

	if xa.Port <= 0 || xa.Port > 65535 {
		return ErrServiceInvalidPort
	}

	return nil
}

type Config struct {
	Env      string `yaml:"env"`
	App      app    `yaml:"app"`
	Server   server `yaml:"server"`
	Services struct {
		XvibeAuth xvibeAuth `yaml:"xvibe-auth"`
	} `yaml:"services"`
}

func New(filepath string) (*Config, error) {
	filepath = strings.TrimSpace(filepath)
	if filepath == "" {
		return nil, ErrInvalidPath
	}

	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	bytes = []byte(os.ExpandEnv(string(bytes)))

	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cfg.Env = strings.TrimSpace(cfg.Env)
	if cfg.Env == "" {
		cfg.Env = "dev"
	}

	cfg.Server.applyDefaults()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if err := c.App.validate(); err != nil {
		return fmt.Errorf("invalid app: %w", err)
	}
	if err := c.Server.validate(); err != nil {
		return fmt.Errorf("invalid server: %w", err)
	}
	if err := c.Services.XvibeAuth.validate(); err != nil {
		return fmt.Errorf("invalid xvibe-auth: %w", err)
	}

	return nil
}
