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
	ErrAppInvalidName    = errors.New("invalid name of app")
	ErrAppInvalidVersion = errors.New("invalid version of app")

	// Server's
	ErrServerInvalidCert           = errors.New("invalid path to server cert")
	ErrServerInvalidKey            = errors.New("invalid path to server key")
	ErrServerTooLittleReadTimeout  = errors.New("too little read timeout")
	ErrServerTooLittleWriteTimeout = errors.New("too little write timeout")
	ErrServerTooLittleIdleTimeout  = errors.New("too little idle timeout")

	// Service's
	ErrServiceInvalidHost       = errors.New("invalid host of service")
	ErrServiceInvalidPort       = errors.New("invalid port of service")
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
		Host string `yaml:"host"` // by def: localhost
		Port int    `yaml:"port"` // by def: 7082
		TLS  struct {
			Enable     bool   `yaml:"enable"`
			ServerCert string `yaml:"server-cert"`
			ServerKey  string `yaml:"server-key"`
		} `yaml:"tls"`
	} `yaml:"http"`
	WS struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"ws"`
	ReadTimeout  time.Duration `yaml:"read-timeout"`
	WriteTimeout time.Duration `yaml:"write-timeout"`
	IdleTimeout  time.Duration `yaml:"idle-timeout"`
}

func (s *server) applyDef() {
	s.HTTP.Host = strings.TrimSpace(s.HTTP.Host)
	if s.HTTP.Host == "" {
		s.HTTP.Host = "localhost"
	}

	if s.HTTP.Port <= 0 || s.HTTP.Port > 65535 {
		s.HTTP.Port = 7082
	}

	s.WS.Host = strings.TrimSpace(s.WS.Host)
	if s.WS.Host == "" {
		s.WS.Host = "localhost"
	}

	if s.WS.Port <= 0 || s.WS.Port > 65535 {
		s.WS.Port = 7083
	}
}

func (s *server) validate() error {
	if s.HTTP.TLS.Enable {
		s.HTTP.TLS.ServerCert = strings.TrimSpace(s.HTTP.TLS.ServerCert)
		if s.HTTP.TLS.ServerCert == "" {
			return ErrServerInvalidCert
		}

		s.HTTP.TLS.ServerKey = strings.TrimSpace(s.HTTP.TLS.ServerKey)
		if s.HTTP.TLS.ServerKey == "" {
			return ErrServerInvalidKey
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

type xvibeMessage struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	TLS  struct {
		Enable     bool   `yaml:"enable"`
		ClientCert string `yaml:"client-cert"`
		ClientKey  string `yaml:"client-key"`
		CaCert     string `yaml:"ca-cert"`
	} `yaml:"tls"`
}

func (xm *xvibeMessage) validate() error {
	xm.Host = strings.TrimSpace(xm.Host)
	if xm.Host == "" {
		return ErrServiceInvalidHost
	}

	if xm.Port <= 0 || xm.Port > 65535 {
		return ErrServiceInvalidPort
	}

	if xm.TLS.Enable {
		xm.TLS.ClientCert = strings.TrimSpace(xm.TLS.ClientCert)
		if xm.TLS.ClientCert == "" {
			return ErrServiceInvalidClientCert
		}

		xm.TLS.ClientKey = strings.TrimSpace(xm.TLS.ClientKey)
		if xm.TLS.ClientKey == "" {
			return ErrServiceInvalidClientKey
		}

		xm.TLS.CaCert = strings.TrimSpace(xm.TLS.CaCert)
		if xm.TLS.CaCert == "" {
			return ErrServiceInvalidCaCert
		}
	}

	return nil
}

type Config struct {
	Env      string `yaml:"env"`
	App      app    `yaml:"app"`
	Server   server `yaml:"server"`
	Services struct {
		XVBMessage xvibeMessage `yaml:"xvibe-message"`
	} `yaml:"services"`
}

// it is implied that .env is loaded
func New(filepath string) (*Config, error) {
	filepath = strings.TrimSpace(filepath)
	if filepath == "" {
		return nil, ErrInvalidPath
	}

	bytes, err := os.ReadFile(os.Getenv("PATH_CONFIG"))
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}
	bytes = []byte(os.ExpandEnv(string(bytes)))

	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	cfg.Env = strings.TrimSpace(cfg.Env)
	if cfg.Env == "" {
		cfg.Env = "dev"
	}

	cfg.Server.applyDef()

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
	if err := c.Services.XVBMessage.validate(); err != nil {
		return fmt.Errorf("invalid xvibe-message: %w", err)
	}

	return nil
}
