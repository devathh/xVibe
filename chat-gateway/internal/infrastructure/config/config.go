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
	ErrServerTooLittleRead  = errors.New("too little read timeout")
	ErrServerTooLittleWrite = errors.New("too little write timeout")
	ErrServerTooLittleIdle  = errors.New("too little idle timeout")
	ErrServerInvalidCert    = errors.New("invalid path to server cert")
	ErrServerInvalidKey     = errors.New("invalid path to server key")
	ErrServerInvalidCaCert  = errors.New("invalid path to ca cert")

	// Xvibe chat's settings
	ErrXvibeChatInvalidHost       = errors.New("invalid host")
	ErrXvibeChatInvalidPort       = errors.New("invalid port")
	ErrXvibeChatInvalidClientCert = errors.New("invalid path to client cert")
	ErrXvibeChatInvalidClientKey  = errors.New("invalid path to client key")

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
			CaCert     string `yaml:"ca-cert"`
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
		s.HTTP.Port = 7081
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

		s.HTTP.TLS.CaCert = strings.TrimSpace(s.HTTP.TLS.CaCert)
		if s.HTTP.TLS.CaCert == "" {
			return ErrServerInvalidCaCert
		}
	}

	if s.ReadTimeout < 100*time.Millisecond {
		return ErrServerTooLittleRead
	}

	if s.WriteTimeout < 100*time.Millisecond {
		return ErrServerTooLittleWrite
	}

	if s.IdleTimeout < 100*time.Millisecond {
		return ErrServerTooLittleIdle
	}

	return nil
}

type xvibeChat struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	TLS  struct {
		Enable     bool   `yaml:"enable"`
		ClientCert string `yaml:"client-cert"`
		ClientKey  string `yaml:"client-key"`
		CaCert     string `yaml:"ca-cert"`
	} `yaml:"tls"`
}

func (xc *xvibeChat) validate() error {
	xc.Host = strings.TrimSpace(xc.Host)
	if xc.Host == "" {
		return ErrXvibeChatInvalidHost
	}

	if xc.Port <= 0 || xc.Port > 65535 {
		return ErrXvibeChatInvalidPort
	}

	if xc.TLS.Enable {
		xc.TLS.ClientCert = strings.TrimSpace(xc.TLS.ClientCert)
		if xc.TLS.ClientCert == "" {
			return ErrXvibeChatInvalidClientCert
		}

		xc.TLS.ClientKey = strings.TrimSpace(xc.TLS.ClientKey)
		if xc.TLS.ClientKey == "" {
			return ErrXvibeChatInvalidClientKey
		}
	}

	return nil
}

type Config struct {
	Env      string `yaml:"env"`
	App      app    `yaml:"app"`
	Server   server `yaml:"server"`
	Services struct {
		XvibeChat xvibeChat `yaml:"xvibe-chat"`
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
	if err := c.Services.XvibeChat.validate(); err != nil {
		return fmt.Errorf("invalid xvibe-chat: %w", err)
	}

	return nil
}
