package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

const (
	minTimeout = 100 * time.Millisecond
)

var (
	// App's
	ErrAppInvalidName    = errors.New("invalid name of service")
	ErrAppInvalidVersion = errors.New("invalid version of service")

	// Server's
	ErrServerInvalidCert           = errors.New("invalid path to cert")
	ErrServerInvalidKey            = errors.New("invalid path to key")
	ErrServerTooLittleReadTimeout  = errors.New("too little read-timeout")
	ErrServerTooLittleWriteTimeout = errors.New("too little write-timeout")
	ErrServerTooLittleIdleTimeout  = errors.New("too little idle-timeout")

	// General's
	ErrInvalidPath      = errors.New("invalid path to config file")
	ErrInvalidTargetURL = errors.New("invalid target url")

	// Rate limit's
	ErrRLTooLittleWindow = errors.New("too little window")
	ErrRLTooLittleLimit  = errors.New("too little limit")
)

type App struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func (a *App) Validate() error {
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

type Server struct {
	HTTP struct {
		Host string `yaml:"host"` // by def: localhost
		Port int    `yaml:"port"` // by def: 8000
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

func (s *Server) SetDefaults() {
	s.HTTP.Host = strings.TrimSpace(s.HTTP.Host)
	if s.HTTP.Host == "" {
		s.HTTP.Host = "localhost"
	}

	if s.HTTP.Port <= 0 || s.HTTP.Port > 65535 {
		s.HTTP.Port = 8000
	}
}

func (s *Server) Validate() error {
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

	if s.ReadTimeout < minTimeout {
		return ErrServerTooLittleReadTimeout
	}
	if s.WriteTimeout < minTimeout {
		return ErrServerTooLittleWriteTimeout
	}
	if s.IdleTimeout < minTimeout {
		return ErrServerTooLittleIdleTimeout
	}

	return nil
}

type RateLimit struct {
	Window time.Duration `yaml:"window"`
	Limit  int           `yaml:"limit"`
}

func (rl *RateLimit) Validate() error {
	if rl.Window < minTimeout {
		return ErrRLTooLittleWindow
	}

	if rl.Limit <= 0 {
		return ErrRLTooLittleLimit
	}

	return nil
}

type Config struct {
	Env     string `yaml:"env"` // by def: dev
	App     App    `yaml:"app"`
	Server  Server `yaml:"server"`
	Service struct {
		RateLimit RateLimit `yaml:"rate-limit"`
	} `yaml:"service"`
	Secrets struct {
		TargetURL string `yaml:"target-url"`
	} `yaml:"secrets"`
}

func (c *Config) SetDefaults() {
	c.Env = strings.TrimSpace(c.Env)
	if c.Env == "" {
		c.Env = "dev"
	}

	c.Server.SetDefaults()
}

func (c *Config) Validate() error {
	if err := c.App.Validate(); err != nil {
		return fmt.Errorf("invalid app: %w", err)
	}
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("invalid server: %w", err)
	}
	if err := c.Service.RateLimit.Validate(); err != nil {
		return fmt.Errorf("invalid rate-limit: %w", err)
	}

	parsedURL, err := url.Parse(c.Secrets.TargetURL)
	if err != nil {
		return fmt.Errorf("invalid target-url: %w", err)
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return ErrInvalidTargetURL
	}

	return nil
}

// New loads and validates configuration
// from the given file path
func New(filepath string) (*Config, error) {
	bytes, err := loadConfigBytes(filepath)
	if err != nil {
		return nil, err
	}

	cfg, err := parseConfig(bytes)
	if err != nil {
		return nil, err
	}

	cfg.SetDefaults()
	return cfg, cfg.Validate()
}

// reading bytes of config file
func parseConfig(bytes []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// loading config's bytes
// from file
func loadConfigBytes(filepath string) ([]byte, error) {
	filepath = strings.TrimSpace(filepath)
	if filepath == "" {
		return nil, ErrInvalidPath
	}

	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	bytes = []byte(os.ExpandEnv(string(bytes)))

	return bytes, nil
}
